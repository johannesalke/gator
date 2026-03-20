## Gator: A RSS feed aggregator
This ReadMe has 2 main sections. The first describes it as a guided project under Boot.dev and what I learned/practiced it. The second describes it as a tool in its own right and fulfills the function of a typical ReadMe.







### Gator as a Project

The purpose of this project was to have students learn how to build a command line tility with variable numbers of arguments, as well as running a Postgres database, managing migrations with goose and generating queery functions with sqlc. 

To a degree it is an extension of the previous project, in which students built a REPL style CLI tool that took the role of a client to make API requests, store the result in memory, and respond to the user based on these replies. 



### Gator as a Tool

Gator is a Command Line Utility that lets users register RSS feeds to follow and automatically sends out requests for updates on those feeds at certain intervalls. 

Requirements: Go toolchain, PostgreSQL

Installation: To install, use the command `go install https://github.com/johannesalke/gator`

Set up: Create a `.gatorconfig.json` file inside your home directory. It's contents should be:
```
{"db_url": |contact string of your postgres database (with sslmode disabled)|,
"current_user_name": "placeholder"
}
```



Commands:
- gator register |name| - Register a new user
- gator login |name| - Log in as this user. No password required.
- gator follow |url| - Follow the feed specified by this url.
- 