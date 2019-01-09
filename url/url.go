package url

import (
  "math/rand"
  "time"
  "net/url"
  "fmt"
)

const (
  size = 5
  symbols = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_-+"
)

type Url struct {
  Id      string      `json:"id"`
  Creation time.Time  `json:"creation"`
  Target string       `json:"target"`
}

type Stats struct {
  Url     *Url  `json:"url"`
  Clicks  int   `json:"clicks"`
}

type Repository interface {
  IdExists(id string) bool
  FindById(id string) *Url
  FindByUrl(url string) *Url
  Save(url Url) error
  RegisterClick(id string)
  FetchClickStats(id string) int
}

var repo Repository


func init() {
  rand.Seed(time.Now().UnixNano())
}


func SetRepo(r Repository) {
  repo = r
}

func FindOrCreateUrl(target string) (u *Url, n bool, err error) {

  if u = repo.FindByUrl(target); u != nil {
    fmt.Println("URI already recorded:", u.Target)
    return u, false, nil
  }

  if _, err = url.ParseRequestURI(target); err != nil {
    fmt.Println("Error:", err)
    return nil, false, err
  }

  url := Url{generateId(), time.Now(), target}
  //fmt.Println("New URI recorded:", url.Target)
  repo.Save(url)
  return &url, true, nil
}

func generateId () string {

  newId := func() string {
    id  := make([]byte, size, size)

    for i := range id {
      id[i] = symbols[rand.Intn(len(symbols))]
    }

    return string(id)
  }

  for {
    if id := newId(); !repo.IdExists(id) {
      return id
    }
  }
}

func Find(id string) *Url {
  return repo.FindById(id)
}

func RegisterClick(id string) {
  repo.RegisterClick(id)
}

func FetchClicks(id string) int {
  return repo.FetchClickStats(id)
}

func (u *Url) Stats() *Stats {
  clicks := repo.FetchClickStats(u.Id)
  return &Stats{u, clicks}
}

