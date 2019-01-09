package main

import (
  "fmt"
  "log"
  "net/http"
  "strings"
  "./url"
  "encoding/json"
  "flag"
)

// Initializing variables that will be accessible by all functions inside this
// package
var (
  port *int // because function in the `flag` package return a pointer
  logMode *bool // Adding logMode
  baseUrl string
  //stats chan string // Deprecated, see Forwarder struct
  repo url.Repository
)

type Headers map[string]string

type Forwarder struct{
  stats chan string
}

// One package may define many different `init` functions. `init` functions are
// useful when the package is distributed into many files. The calling order of
// those functions is undefined.
func init() {
  port = flag.Int("p", 1234, "port")
  logMode = flag.Bool("l", true, "log on/off")

  flag.Parse()

  baseUrl = fmt.Sprintf("http://0.0.0.0:%d", *port)
  repo = url.CreateMemoryRepo()
}

func main() {

  stats := make(chan string)
  defer close(stats)
  go RegisterClick(stats)

  url.SetRepo(repo)
  http.HandleFunc("/api/shorten", Shortener)
  http.HandleFunc("/api/stats/", StatsReporter)
  // Before refactoring it looked like this
  //http.HandleFunc("/go.to/", Forwarder)
  // And now we are be able to inject the stats channel into the Forwarder
  http.Handle("/go.to/", &Forwarder{stats})

  myLog("Listening on %s", baseUrl)
  log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *port), nil))
}

func Shortener(w http.ResponseWriter, r *http.Request) {

  if r.Method != "POST" {
    respondWith(w, http.StatusMethodNotAllowed, Headers{"Allow": "POST"})
    return
  }

  url, new_url, err := url.FindOrCreateUrl(extractUrl(r))
  myLog("URL object: '%+v', is it new? '%t', errors: '%+v'", url, new_url, err)

  if err != nil {
    respondWith(w, http.StatusBadRequest, nil)
    return
  }

  var status int
  if new_url {
    status = http.StatusCreated
  } else {
    status = http.StatusOK
  }

  shortUrl := fmt.Sprintf("%s/go.to/%s", baseUrl, url.Id)
  respondWith(w, status, Headers{
    "Location": shortUrl,
    "Link": fmt.Sprintf("<%s/api/stats/%s>; rel=\"stats\"", baseUrl, url.Id),
  })

  myLog("URL '%s' successfully shortened to '%s'.", url.Target, shortUrl)
}

func respondWith(w http.ResponseWriter, status int, headers Headers) {
  for k,v := range headers {
    w.Header().Set(k,v)
  }
  w.WriteHeader(status)
}

func extractUrl(r *http.Request) string {
  url := make([]byte, r.ContentLength, r.ContentLength)
  r.Body.Read(url)

  return string(url)
}

func findUrlAndExecute (
  w http.ResponseWriter,
  r *http.Request,
  runFunc func(*url.Url),
) {

  path := strings.Split(r.URL.Path, "/")
  id := path[len(path)-1]
  myLog("Looking for ID: '%s'", id)

  if url := url.Find(id); url != nil {

    myLog("Found URI: '%s'", url.Target)
    runFunc(url)

  } else {

    myLog("Could not find URI under ID: '%s'", id)
    http.NotFound(w, r)
  }
}

// Before refactoring this function signature was like this:
//func Forwarder(w http.ResponseWriter, r *http.Request) {
// But now it looks like this, and it is because we're trying to reduce the
// amount of global variables.
// After the second refactor round some code duplication was removed, and this
// is the new body:
func (red *Forwarder) ServeHTTP (w http.ResponseWriter, r *http.Request) {

  redirectIfFound := func(url *url.Url) {
    http.Redirect(w, r, url.Target, http.StatusMovedPermanently)
    red.stats <- url.Id
  }

  // After refactoring:
  findUrlAndExecute(w, r, redirectIfFound)
}

func StatsReporter(w http.ResponseWriter, r *http.Request) {
  jsonResponder := func(url *url.Url) {
    json, err := json.Marshal(url.Stats())

    if err != nil {
      w.WriteHeader(http.StatusInternalServerError)
      return
    }

    respondWithJson(w, string(json))
  }

  // After refactoring:
  findUrlAndExecute(w, r, jsonResponder)
}

func RegisterClick(ids <-chan string) {
  for id := range ids {
    url.RegisterClick(id)
    // This one uses the `FetchClicks` from the url package
    //fmt.Printf("%d clicks for %s\n", url.FetchClicks(id), id)

    // And this one uses the method declared for the struct repo, that
    // implicitly implements the Repository interface
    myLog("%d clicks for '%s'.", repo.FetchClickStats(id), id)
  }
}

func respondWithJson(w http.ResponseWriter, answer string) {
  respondWith(w, http.StatusOK, Headers{
    "Content-Type": "application/json",
  })

  fmt.Fprintf(w, answer)
}

func myLog(format string, values ...interface{}) {
  if *logMode {
    log.Printf(fmt.Sprintf("%s\n", format), values...)
  }
}

