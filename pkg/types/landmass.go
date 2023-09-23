package types

type Landmass struct {
	ProjectID string `db:"project_id"`
	ID        string `db:"id"`
	Epoch     int    `db:"epoch"`
	Size      int    `db:"size"`

	ColorR int `db:"color_r"`
	ColorG int `db:"color_g"`
	ColorB int `db:"color_b"`

	FirstX int `db:"first_x"`
	FirstY int `db:"first_y"`
}
