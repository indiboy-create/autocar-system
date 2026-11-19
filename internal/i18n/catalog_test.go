package i18n

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"testing"
)

// goFiles 遍历应用各包的非测试 Go 文件；skipI18n 为真时跳过本包。
func goFiles(t *testing.T, skipI18n bool, visit func(path string, f *ast.File, fset *token.FileSet)) {
	t.Helper()
	roots := []string{filepath.Join("..", "..", "internal"), filepath.Join("..", "..", "main.go")}
	for _, root := range roots {
		err := filepath.WalkDir(root, func(path string, e fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if e.IsDir() {
				if skipI18n && path == filepath.Join("..", "..", "internal", "i18n") {
					return fs.SkipDir
				}
				return nil
			}
			if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
				return nil
			}
			fset := token.NewFileSet()
			f, err := parser.ParseFile(fset, path, nil, 0)
			if err != nil {
				return err
			}
			visit(path, f, fset)
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}
}

// markedKey 返回被 i18n 标记函数标注的文案与是否需要译文。
// T/Tf/Errorf/N 要求字典有译文；Data 是不展示的原文数据，无需译文。
func markedKey(call *ast.CallExpr) (key string, needsTranslation bool, ok bool) {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return "", false, false
	}
	switch sel.Sel.Name {
	case "T", "Tf", "Errorf", "N":
		needsTranslation = true
	case "Data":
		needsTranslation = false
	default:
		return "", false, false
	}
	pkg, ok := sel.X.(*ast.Ident)
	if !ok || pkg.Name != "i18n" || len(call.Args) == 0 {
		return "", false, false
	}
	text, ok := stringLiteral(call.Args[0])
	return text, needsTranslation, ok
}

func stringLiteral(e ast.Expr) (string, bool) {
	lit, ok := e.(*ast.BasicLit)
	if !ok || lit.Kind != token.STRING {
		return "", false
	}
	s, err := strconv.Unquote(lit.Value)
	return s, err == nil
}

var formatVerb = regexp.MustCompile(`%(\[\d+\])?[-+# 0-9.]*[a-zA-Z]`)

// verbs 返回文案里的格式动词，去掉参数序号后排序，忽略词序差异。
func verbs(s string) string {
	var out []string
	for _, v := range formatVerb.FindAllString(s, -1) {
		v = regexp.MustCompile(`\[\d+\]`).ReplaceAllString(v, "")
		out = append(out, v)
	}
	sort.Strings(out)
	return strings.Join(out, " ")
}

// markedLiterals 收集文件里所有被 i18n 标记的字面量位置。
func markedLiterals(f *ast.File) map[token.Pos]bool {
	out := map[token.Pos]bool{}
	ast.Inspect(f, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		if _, _, ok := markedKey(call); ok && len(call.Args) > 0 {
			if lit, ok := call.Args[0].(*ast.BasicLit); ok {
				out[lit.Pos()] = true
			}
		}
		return true
	})
	return out
}

// TestCatalogHasEveryKey 保证每个被翻译的文案都有译文，且格式动词一致：
// 否则英文模式下会漏出中文，或 Sprintf 因占位符不匹配而输出错误内容。
func TestCatalogHasEveryKey(t *testing.T) {
	seen := 0
	goFiles(t, true, func(path string, f *ast.File, fset *token.FileSet) {
		ast.Inspect(f, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			key, need, ok := markedKey(call)
			if !ok || !need {
				return true
			}
			seen++
			en, ok := dict[key]
			if !ok {
				t.Errorf("%s: no translation of %q", fset.Position(call.Pos()), key)
				return true
			}
			if verbs(key) != verbs(en) {
				t.Errorf("%s: translation of %q has other verbs: %q", fset.Position(call.Pos()), key, en)
			}
			return true
		})
	})
	if seen < 100 {
		t.Errorf("found only %d translated texts: the scan has gone wrong", seen)
	}
}

// TestNoHardcodedChinese 禁止 i18n 包外出现未标记的中文字面量：
// 它们会绕过语言切换，在英文模式下依然显示中文。
func TestNoHardcodedChinese(t *testing.T) {
	han := regexp.MustCompile(`\p{Han}`)
	goFiles(t, true, func(path string, f *ast.File, fset *token.FileSet) {
		marked := markedLiterals(f)
		ast.Inspect(f, func(n ast.Node) bool {
			lit, ok := n.(*ast.BasicLit)
			if !ok || lit.Kind != token.STRING || marked[lit.Pos()] {
				return true
			}
			if han.MatchString(lit.Value) {
				t.Errorf("%s: hardcoded Chinese %s", fset.Position(lit.Pos()), lit.Value)
			}
			return true
		})
	})
}
