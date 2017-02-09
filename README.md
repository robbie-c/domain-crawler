
### Overview
This is a web crawler written in Go that will crawl all pages on a particular domain.
It will output all pages and CSS files it finds, as well as what resources they link to.

### TODO List
* Respect robots.txt
* Seed with sitemap.xml
* Print indirect resources (e.g. page that points to CSS file that points to font, which right now involves looking at two entries on the )
* Write more tests!
* Move the cmd utility and the core code to separate repos