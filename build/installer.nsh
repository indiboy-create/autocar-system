; mygo's installer defaults to %LOCALAPPDATA%\Programs\<slug> and remembers a
; user-chosen directory via InstallDirRegKey, so we deliberately do NOT force
; $INSTDIR here. This keeps the app off any fixed drive and lets upgrades stay
; in place. data/ and media/ still live next to the executable.
;
; On uninstall the data is preserved under %APPDATA% (the user's roaming
; profile, independent of the install drive) and restored on the next install.

!define PRESERVED "$APPDATA\autoCarSystem-preserved"

!macro customInstall
  CreateDirectory "$INSTDIR\data"
  CreateDirectory "$INSTDIR\media"
  CreateDirectory "$PRESERVED\data"
  CreateDirectory "$PRESERVED\media"

  IfFileExists "$PRESERVED\data\*" 0 +2
    CopyFiles /SILENT "$PRESERVED\data\*" "$INSTDIR\data"

  IfFileExists "$PRESERVED\media\*" 0 +2
    CopyFiles /SILENT "$PRESERVED\media\*" "$INSTDIR\media"
!macroend

!macro customUnInstall
  CreateDirectory "$PRESERVED\data"
  CreateDirectory "$PRESERVED\media"

  IfFileExists "$INSTDIR\data\*" 0 +2
    CopyFiles /SILENT "$INSTDIR\data\*" "$PRESERVED\data"

  IfFileExists "$INSTDIR\media\*" 0 +2
    CopyFiles /SILENT "$INSTDIR\media\*" "$PRESERVED\media"
!macroend
