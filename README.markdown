A simple start page with a background image from [/r/EarthPorn](http://www.reddit.com/r/earthporn), weather from [yr.no](http://www.yr.no) and customizable links.

## Screenshot
![Screenshot](http://i.imgur.com/u42QOZe.png)

## Installation

`go get github.com/kch42/startpage`

## Configuration

The startpage configuration is located in the file ~/.startpagerc. It is a list of commands. A command has a name and can optionally have parameters separated by spaces or tabs. A backspace `\\` will interpret the next charcter literally (can be used to escape whitespace, linebreaks and backspaces). Commands are separated by newlines.

These commands are implemented:

### `set-weather-place`

Takes one argument, the place used for weather info. startpage uses [yr.no](http://www.yr.no) to get weather data. Use the search box on that page to search for your place. You will then be redirected to an URL like this: `http://www.yr.no/place/<myplace>`. Put the `<myplace>` part after the `set-weather-place` command like this:

	set-weather-place <myplace>

### `add-link`

Add a link that is displayed on the startpage. First argument is the title, second one the URL.

Example:

	add-link github           http://www.github.com
	add-link reddit           http://www.reddit.com
	add-link go               http://www.golang.org
	add-link another\ example http://www.example.org

## Running

If `$GOPATH/bin` is in your `$PATH`, you can run startpage with the command `startpage`. By default, startpage listens on port 25145. You can change that with a command line switch: `startpage -laddr :<port>`