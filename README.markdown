A simple start page with a background image from [/r/EarthPorn](http://www.reddit.com/r/earthporn), weather from [yr.no](http://www.yr.no) and customizable links.

## Screenshot
![Screenshot](http://i.imgur.com/u42QOZe.png)

## Installation

`go get github.com/kch42/startpage`

## Configuration

startpage uses two files in your home directory for configuration:

### ~/.startpage-urls

This describes the hyperlinks that are displayed on the startpage. A list of key-value-pairs. Each line is such a pair. Key and value are separated with `->`. The key is the title of the link, the value the URL.

Example:

	github -> http://www.github.com
	reddit -> http://www.reddit.com
	go -> http://www.golang.org
	example -> http://www.example.org

### ~/.startpage-weather

The place for the weather is stored here. startpage uses [yr.no](http://www.yr.no) to get weather data. Use the search box on that page to search for your place. You will then be redirected to an URL like this: `http://www.yr.no/place/<myplace>`. Put the `<myplace>` part into the `.startpage-weather` file.

## Running

If `$GOPATH/bin` is in your `$PATH`, you can run startpage with the command `startpage`. By default, startpage listens on port 25145. You can change that with a command line switch: `startpage -laddr :<port>`