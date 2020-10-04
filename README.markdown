A simple start page with a background image from a subreddit ([/r/EarthPorn](http://www.reddit.com/r/earthporn) by default), weather from [yr.no](http://www.yr.no) and customizable links.

## Screenshot
![Screenshot](http://i.imgur.com/u42QOZe.png)

## Installation

`go get github.com/silvasur/startpage`

## Configuration

The optional startpage configuration is a JSON file located at `~/.config/startpage/config.json`.

Here is an example with all fields filled out.

    {
        // The place for which to get the weather data. If omitted, no weather will be shown
        "WeatherPlace": "Germany/Hamburg/Hamburg",

        // A list of links to show. Can be omitted.
        "Links": [
            {
                "Title": "example",
                "URL": "https://www.example.com"
            }
        ],

        // If set, background images can be saved here
        "BackgroundSavepath": "/home/laria/Pictures/cool-backgrounds",

        // If set, this limits the background image size, the default is DEFAULT_BACKGROUND_MAXDIM (=2500)
        "BackgroundMaxdim": 4000,

        // Get background images from this subreddit. Defaults to "EarthPorn"
        "ImageSubreddit": "ruralporn"
    }

## Running

If `$GOPATH/bin` is in your `$PATH`, you can run startpage with the command `startpage`. By default, startpage listens on port 25145. You can change that with a command line switch: `startpage -laddr :<port>`
