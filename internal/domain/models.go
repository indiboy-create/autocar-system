package domain

type Category struct {
	ID      int64
	Name    string
	Sort    int
	Enabled bool
}

type Yuvision struct {
	ID          int64
	BrandID     int64
	FirmID      int64
	SeriesID    int64
	CategoryID  int64
	Name        string
	CoverImage  string
	Images      string
	Description string
	CreatedAt   string
	UpdatedAt   string
}

type Gallery struct {
	ID          int64
	YouvisionID int64
	BrandID     int64
	FirmID      int64
	SeriesID    int64
	CategoryID  int64
	Title       string
	CoverImage  string
	Images      string
	ImageCount  int
	CreatedAt   string
	UpdatedAt   string
}
