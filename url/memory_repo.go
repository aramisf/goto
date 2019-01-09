package url

type memoryRepo struct {
  urls map[string]*Url
  clicks map[string]int
}

func CreateMemoryRepo() *memoryRepo {
  return &memoryRepo{
    make(map[string]*Url),
    make(map[string]int),
  }
}

func (r *memoryRepo) IdExists(id string) bool {
  _, exists := r.urls[id] // because a map returns two values: the first is the
                          // value itself and the second is a boolean indicating
                          // that the searched key DOES exist
  return exists
}

func (r *memoryRepo) FindById(id string) *Url {
  return r.urls[id]
}

// r.urls is a map indexed by IDs, which values are URIs, that's why one must
// loop through the keys and match against its values in order to find a
// matching URI
func (r *memoryRepo) FindByUrl(url string) *Url {
  for _, u := range r.urls {
    if u.Target == url {
      return u
    }
  }

  return nil
}

func (r *memoryRepo) Save(url Url) error {
  r.urls[url.Id] = &url
  return nil
}

func (r *memoryRepo) RegisterClick(id string) {
  r.clicks[id] += 1
}

func (r *memoryRepo) FetchClickStats(id string) int {
  return r.clicks[id]
}

