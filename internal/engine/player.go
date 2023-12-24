package engine

type Player struct {
	Id    int32
	Name  string
	Score int32
}

func NewPlayer(id int32, name string, score int32) *Player {
	return &Player{
		Id:    id,
		Name:  name,
		Score: score,
	}
}
