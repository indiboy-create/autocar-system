package domain

type Brand struct {
	ID             int64
	Name           string
	FirstLetter    string
	LogoURL        string
	LogoLocalPath  string
	State          int64
	HaveNewEnergy  int64
	APISort        int64
	CollectedAt    string
	UserUpdatedAt  string
	MediaUpdatedAt string
}

type BrandEdit struct {
	Name          string
	FirstLetter   string
	LogoURL       string
	LogoLocalPath string
	State         int64
	HaveNewEnergy int64
	APISort       int64
}

type Firm struct {
	ID             int64
	BrandID        int64
	Name           string
	PY             string
	Link           string
	LogoURL        string
	LogoLocalPath  string
	CollectedAt    string
	UserUpdatedAt  string
	MediaUpdatedAt string
}

type FirmEdit struct {
	Name          string
	PY            string
	Link          string
	LogoURL       string
	LogoLocalPath string
}

type Series struct {
	ID             int64
	BrandID        int64
	FirmID         int64
	Name           string
	GroupName      string
	State          int64
	PriceMin       float64
	PriceMax       float64
	Place          string
	PY             string
	LogoURL        string
	LogoLocalPath  string
	LevelName      string
	NewEnergy      int64
	IsForeignCar   int64
	IsTop          int64
	FuelTypes      string
	CartypeName    string
	Tag1           string
	Rank           int64
	UVRank         int64
	HotRank        int64
	CollectedAt    string
	UserUpdatedAt  string
	MediaUpdatedAt string
}

type SeriesEdit struct {
	Name          string
	GroupName     string
	State         int64
	PriceMin      float64
	PriceMax      float64
	Place         string
	PY            string
	LogoURL       string
	LogoLocalPath string
	LevelName     string
	NewEnergy     int64
	IsForeignCar  int64
	IsTop         int64
	FuelTypes     string
	CartypeName   string
	Tag1          string
	Rank          int64
	UVRank        int64
	HotRank       int64
}

type Spec struct {
	ID                  int64
	BrandID             int64
	FirmID              int64
	SeriesID            int64
	GroupName           string
	Name                string
	SeriesName          string
	ModelYear           string
	BodyForm            string
	PriceMin            float64
	PriceMax            float64
	NewsPrice           float64
	NewsID              int64
	DealerID            int64
	SaleType            string
	ImageURL            string
	ImageLocalPath      string
	ImageSHA256         string
	WhiteImageURL       string
	WhiteImageLocalPath string
	WhiteImageSHA256    string
	CollectedAt         string
	UserUpdatedAt       string
	MediaUpdatedAt      string
}

type SpecEdit struct {
	GroupName           string
	Name                string
	SeriesName          string
	ModelYear           string
	BodyForm            string
	PriceMin            float64
	PriceMax            float64
	NewsPrice           float64
	NewsID              int64
	DealerID            int64
	SaleType            string
	ImageURL            string
	ImageLocalPath      string
	ImageSHA256         string
	WhiteImageURL       string
	WhiteImageLocalPath string
	WhiteImageSHA256    string
}
