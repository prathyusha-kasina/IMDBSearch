# IMDBSearch
Search engine for 1000 top imdb movies

Functionality supported:
Search based on director name and genre, would return a list of movies


APIs and Queries are supported:
GET on / shows the usage.
GET on /search?category=director&name=Christopher+Nolan returns a list of movies directed by Christopher Nolan.
GET on /search?category=genre&name=Action returns a list of movies whose genre is Action.

Algorithm used:
1.Get list of top 1000 movie urls
2.Scrape each movie's html to extract movie metadata
3.populate hash maps with the data
4.Develop APIs to support search

Steps to execute:
1.Create a folder with src, bin, pkg.
2.set the path to the folder as GOPATH.
3.git clone IMDBSearch in the src folder.
4.run "go get github.com/gocolly/colly"
5.run "go get github.com/gorilla/mux"
6.cd to IMDBSearch and run "go install"
7.A binary named IMDBSearch is created in bin folder 
8.Execute it.
9.Right now populating maps is taking 1 min.Please wait till then.
10.Open a browser and give the following URLs
http://localhost:8080/
http://localhost:8080/search?category=director&name=directorname
http://localhost:8080/search?category=genre&name=genrename
