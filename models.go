package main

// Torrent holds metadata for a single torrent listing.
type Torrent struct {
	ID        int    `json:"id"`
	Title     string `json:"title"`
	Category  string `json:"category"`
	Uploaded  string `json:"uploaded"`
	Seeders   int    `json:"seeders"`
	Leechers  int    `json:"leechers"`
	Completed int    `json:"completed"`
	Size      string `json:"size"`
	File      string `json:"file"`
	Link      string `json:"link"`
	Magnet    string `json:"magnet"`
}

// Comment is a single user comment on a torrent page.
type Comment struct {
	Name      string `json:"name"`
	Content   string `json:"content"`
	Image     string `json:"image"`
	Timestamp string `json:"timestamp"`
}

// Comments wraps a comment count and list.
type Comments struct {
	Count    int       `json:"count"`
	Comments []Comment `json:"comments"`
}

// File is the detailed view of a single torrent.
type File struct {
	Torrent     Torrent  `json:"torrent"`
	Description string   `json:"description"`
	SubmittedBy string   `json:"submittedBy"`
	InfoHash    string   `json:"infoHash"`
	CommentInfo Comments `json:"commentInfo"`
}

// QueryParams holds optional search/filter parameters.
type QueryParams struct {
	Query  string
	Sort   string
	Order  string
	Page   string
	Filter string
}
