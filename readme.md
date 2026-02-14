# Short-URL
A very basic URL shortener.

I started teaching myself [Go](https://go.dev) a few days ago, and wanted a simple exercise that would still give me a bit of a challenge. While this is simplistic, it's still a usable piece of software. I wouldn't recommend running it with external access though since there's no user accounts or security.

Since one of the first questions anyone asks these days is predictable: no, this is not vibe coded. I did not use AI at all. I'm learning Golang


## Features
 * Basic API usage (GET/POST/DELETE)
 * Slugs stored in memory or SQLite
 * Readme with curl examples
 * JSON logging
 * JSON responses
 * Docker build


## Getting Started
For now, run the application locally with whatever port you want to use:
```bash
go run ./src -port 1234
```

By default the application uses a simple memory store. If you want persistent storage, set the database driver to `sqlite`:
```bash
go run ./src -port 1234 -database sqlite
```


Then open another terminal window and test with `curl` like this:
```bash
curl -X POST -d "link=https://www.google.com" http://localhost:4000/shorten
```

Note the default port; you'll have to replace it with whatever you choose. But what you're doing is using curl with:
 * `-X POST` makes the call using the POST verb
 * `-d "link=https://etc"` passes form data in the body, using the "link" form field name

This call returns the slug. In the case of this input, it should be `rGu2ae`

Now you can retrieve the value using the slug:
```bash
curl http://localhost:4000/rGu2ae
```

Stop the server with `CTRL`+`C`


### Building the Docker container
The provided multi-stage build will handle building the application and packing it a container (about 91 MB on Debian slim).

First, build the container:
```bash
docker build . -t short-url-app
```

Then run it with appropriate environment variables. You can do that with the provided compose file, or manually with something like:
```bash
docker run -p 8080:9000 -e PORT=9000 -e DATABASE=memory short-url-app
```
(where 8080 would be the port on your machine and 9000 is what the container listens on)

Just remember that if you want to use SQLite you should have a local volume or bind mount so the data can persist:
```bash
mkdir ./data
docker run -p 8080:9000 -e PORT=9000 -e DATABASE=sqlite -v ./data:/app/data short-url-app
```


## License
MIT. Basically do whatever you want with this. Maybe fork it and implement some additional features?


## Future
I don't currently have any plans to implement these ideas, but if you want to fork this repo and play around, I suggest:

* ⚠️ **Safety features** = if you expose this service to the world, it will likely be used for horrific stuff. So add some safety features like letting people flag links to illegal content, comparing links to known "bad" domains, banning accounts that post bad links, etc.<hr>
* **Database support** = MySQL, PostgreSQL, etc
* **More CRUD support** = I implemented POST and GET. You can implement DELETE and PUT.
* **OIDC / User accounts** = let users register and maintain their links
* **User Interface** = Make a UI so users can view and maintain their shortened links
* **Random Slugs** = I chose to make the slugs using SHA hashing and base64 encoding. This is obviously not random. I was focused on generating (mostly) unique slugs that can also be validated with unit tests. Simply replacing the SHA hash's `link` parameter with something unique like a GUID would make this a lot more random.
