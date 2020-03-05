package main

import (
	"fmt"
	"github.com/gocolly/colly"
	"strings"
	"encoding/json"
	"strconv"
	"sync"
        "github.com/gorilla/mux"
	"net/http"
	"log"
)

type Dirdetails struct {
	DirWrapper struct {
		Rawdir json.RawMessage `json:"director"`
	}
	dir  Dir   `json:"-"`
	dirarray []Dir `json:"-"`
}

type Genredetails struct {
	GenreWrapper struct {
		Rawgenre json.RawMessage `json:"genre"`
	}
	genre  string   `json:"-"`
	genrearray []string `json:"-"`
}

type Dir struct {
	Name string `json:"name"`
}

type Otherdetails struct {
	Moviename string `json:"name"`
}

type moviedetails struct{
	director interface{}
	genre interface{}
	moviename string
}

type response struct{
	Message string
	ListOfMovies []string
}

var mutex = &sync.Mutex{}

//this map stores keys as directors and values as list of movies
var  dirmap = make(map[string][]string)

//this map stores keys as genres and values as list of movies
var  genremap = make(map[string][]string)

//this map stores movie names as keys and moviedetails as values
var moviemap = make(map[string]moviedetails)

func main() {

	urls := urlGenerator()
	queue := crawlUrls(urls)
	movieinfo := parseData(queue)
	populateMap(movieinfo, moviemap, genremap, dirmap)

	fmt.Println("data from IMDB is fetched and populated in a map")

	r := mux.NewRouter()
	r.Path("/").HandlerFunc(usage).Methods(http.MethodGet)
	r.Path("/search").Queries("category", "{category}").Queries("name", "{name}").HandlerFunc(search).Methods(http.MethodGet)
	log.Fatal(http.ListenAndServe(":8080", r))
}

//This method is for displaying the usage of the URL and query parameters
//This method is called on a GET on http://localhost:8080/
func usage(w http.ResponseWriter, r *http.Request){
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(getBytes(""))
}

//This method is used for combining error message and usage message
//Created mainly for modularity and code reuse
func getBytes(msg string) []byte {
	defaultMsg := "Welcome to the IMDB search.This api lets you search by providing query parameters.The supported actions are GET on /search?catergory=director&name=directorname and /search?category=genre&name=adventure"
	if strings.Compare(msg,"") != 0 {
		defaultMsg = msg + "\n" + defaultMsg
	}
	resp := response{defaultMsg, []string{}}
	bytes,_ := json.Marshal(resp)
	return bytes
}
//This method checks if query parameters category and name 
//are correctly provided
func checkParams(pathParams map[string]string) ([]string, string) {
	category, ok := pathParams["category"]
	if !ok {
		msg := "category is not provided as query parameter"
		return nil, msg
	}

	if strings.Compare(category,"director")!=0 && strings.Compare(category,"genre")!=0{
                msg := "category can be director or genre"
                return nil, msg
        }
	name := ""
	name, ok = pathParams["name"]
	if !ok {
                msg := "name is not provided as a query parameter"
                return nil, msg
        }
	return []string{category,name},""
}
//This method is invoked by a GET on /search 
//This method validates the query parameters and writes an appropriate
//error message and a list of movies to the response body
func search(w http.ResponseWriter, r *http.Request){
	pathParams := mux.Vars(r)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	params, errMsg := checkParams(pathParams)
	if strings.Compare(errMsg, "") != 0 {
		w.Write(getBytes(errMsg))
		return
	}

	category := params[0]
	name := params[1]

	if strings.Compare(category,"director")==0{
		list,ok := dirmap[name]
		if !ok {
			msg := "not a top director"
			resp := response{msg,nil}
			bytes, _ := json.Marshal(resp)
			w.Write(bytes)
			return
		}
		bytes,_ := json.Marshal(response{"",list})
		w.Write(bytes)
	}

	if strings.Compare(category,"genre")==0{
                list,ok := genremap[name]
                if !ok {
			msg := "not a top genre"
			resp := response{msg,nil}
			bytes,_ := json.Marshal(resp)
                        w.Write(bytes)
			return
                }
		bytes,_ := json.Marshal(response{"",list})
                w.Write(bytes)
        }
}

//This method populates moviemap with movieinfo
func populateMovieMap(oneMovieInfo moviedetails, moviemap map[string]moviedetails){
	if _, ok := moviemap[oneMovieInfo.moviename]; !ok {
		moviemap[oneMovieInfo.moviename] = oneMovieInfo
	}
}

//This method populates genremap with movieinfo
func populateGenreMap(oneMovieInfo moviedetails, genremap map[string][]string) {
	if _, ok := oneMovieInfo.genre.([]string); !ok {
		//genre is a string
		a := oneMovieInfo.genre.(string)
		val, present := genremap[a]
		if present{
			val = append(val, oneMovieInfo.moviename)
			genremap[a] = val
		}else {
			movies := []string{}
			movies = append(movies, oneMovieInfo.moviename)
			genremap[a] = movies
		}
	}else{
		//genre is a list of strings
		a:= oneMovieInfo.genre.([]string)
		for j:=0;j<len(a);j++ {
			val, present := genremap[a[j]]
			if present{
				val = append(val, oneMovieInfo.moviename)
				genremap[a[j]] = val
			}else {
				movies := []string{}
				movies = append(movies, oneMovieInfo.moviename)
				genremap[a[j]] = movies
			}
		}
	}
}

//This method populates dirmap with movieinfo
func populateDirMap(oneMovieInfo moviedetails, dirmap map[string][]string) {
	if _, ok := oneMovieInfo.director.([]Dir); !ok {
		//director is not an array of structs
		a := oneMovieInfo.director.(Dir)
		val, present := dirmap[a.Name]
		if present{
			val = append(val, oneMovieInfo.moviename)
			dirmap[a.Name] = val
		}else {
			movies := []string{}
			movies = append(movies, oneMovieInfo.moviename)
			dirmap[a.Name] = movies
		}
	}else{
		//director is an array of structs Dir
		a:= oneMovieInfo.director.([]Dir)
		for j:=0;j<len(a);j++ {
			val, present := dirmap[a[j].Name]
			if present{
				val = append(val, oneMovieInfo.moviename)
				dirmap[a[j].Name] = val
			}else {
				movies := []string{}
				movies = append(movies, oneMovieInfo.moviename)
				dirmap[a[j].Name] = movies
			}
		}
	}
}
//This method populates list of moviedetails into 3 maps.
//moviemap contains movie name as key and moviedetails as value
//genremap contains genre as key and a list of movies as value
//dirmap contains director name as key and a list of movies as value
func populateMap(movieinfo []moviedetails, moviemap map[string]moviedetails, genremap, dirmap map[string][]string) {
	for i:=0;i<len(movieinfo);i++ {
		m := movieinfo[i]
		populateMovieMap(m, moviemap)
		populateDirMap(m, dirmap)
		populateGenreMap(m, genremap)
	}
}

//This method visits all the individual movie urls by creating 100 go routines at a time 
// and unmarshalls the json
//This method returns a list of structs of type moviedetails
func parseData(queue []string) ([]moviedetails){
	movieinfo := []moviedetails{}
	var ch chan string = make(chan string)
	var count = 0

	for i:=0;i<len(queue);i=i+100 {
		for j:=i;j<i+100 && j<len(queue);j++ {
			str := queue[j]
			go visitUrls(ch, str)
			count++
		}
		for count>0 {
			msg := <- ch
			count--
			md := unmarshalJson([]byte(msg))
			movieinfo = append(movieinfo, md)
		}
	}
	return movieinfo
}

//This method is run in go routine
//It checks each individual movie url and gets the schema under the script tag
//The schema contains director, genre and other info
func visitUrls(ch chan string, url string) {
	c := colly.NewCollector()
        c.OnHTML("head", func(e *colly.HTMLElement) {
                e.ForEach("script", func(_ int, el *colly.HTMLElement) {
                        txt := el.Text
                        if(strings.Contains(txt,"http://schema.org")) {
				ch <- txt
                        }
                })
        })
	c.Visit(url)
}

//This method gets called by json.Unmarshal after the unmarshaler is set
func (r *Dirdetails) UnmarshalJSON(b []byte) error {
	if err := json.Unmarshal(b, &r.DirWrapper); err != nil {
		return err
	}
	if r.DirWrapper.Rawdir[0] == '[' {
		return json.Unmarshal(r.DirWrapper.Rawdir, &r.dirarray)
	}
	return json.Unmarshal(r.DirWrapper.Rawdir, &r.dir)
}

//This method gets called by json.Unmarshal after the unmarshaler is set
func (r *Genredetails) UnmarshalJSON(b []byte) error {
	if err := json.Unmarshal(b, &r.GenreWrapper); err != nil {
		return err
	}
	if r.GenreWrapper.Rawgenre[0] == '[' {
		return json.Unmarshal(r.GenreWrapper.Rawgenre, &r.genrearray)
	}
	return json.Unmarshal(r.GenreWrapper.Rawgenre, &r.genre)
}

//The input to the method is the schema extracted from each movie url
//The genre and director in the schema do not have a fixed type.
//Sometimes they are arrays of structs/strings  and sometimes structs/strings
//So, movie name is extarcted first
//Then unmarshaler is set and extracted
func unmarshalJson(text []byte) moviedetails {
	md := moviedetails{}

	od := Otherdetails{}
	err := json.Unmarshal(text,&od)
	if err!=nil {
		fmt.Println("error unmarshalling into otherdetails")
	}
	md.moviename = od.Moviename

	var _ json.Unmarshaler = &Dirdetails{}
	r := Dirdetails{}
	err = json.Unmarshal(text, &r)
	if err!=nil{
		fmt.Println("error unmarshalling into dirdetails")
	}

	if len(r.dirarray)==0 {
		md.director = r.dir
	}else {
		md.director = r.dirarray
	}

	var _ json.Unmarshaler = &Genredetails{}
	gen := Genredetails{}
	err = json.Unmarshal(text, &gen)
	if err!=nil{
		fmt.Println("error unmarshalling into genredetails")
	}
	if len(gen.genrearray)==0 {
		md.genre = gen.genre
	}else {
		md.genre = gen.genrearray
	}
	return md
}

//This method visits the urls listed in the input and looks for href links in the html and scrapes them
//It further checks if the links contain a certain pattern, and if yes is a valid link to a movie
//This method returns a list of individual movie urls.
func crawlUrls(urls []string) []string {
	queue := []string{}
	c := colly.NewCollector()
	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		if strings.HasPrefix(link, "/title/") && strings.HasSuffix(link, "/?ref_=adv_li_i") && strings.Index(link, "/title/")==0 {
			queue = append(queue,e.Request.AbsoluteURL(link))
		}
	})
	for i:=0;i<len(urls);i++ {
		c.Visit(urls[i])
	}
	return queue
}


//IMDB top 1000 movies are spanned across multiple pages as per pagination.
//So, this method generates the urls of imdb pages containing 100 movies per page
//The number 100 is hardcoded as the website seems to support only certain pagination values
//This method returns an array of urls 
func urlGenerator() []string {
	urls := []string{}
	for startMovieNum := 1; startMovieNum<=901; startMovieNum = startMovieNum + 100 {
		url := "https://www.imdb.com/search/title/?groups=top_1000&view=simple&sort=user_rating,desc&count=100&start=" + strconv.Itoa(startMovieNum) + "&ref_=adv_nxt"
		urls = append(urls,url)
	}
	return urls
}
