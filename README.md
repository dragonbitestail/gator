# Gator

A basic, command-line [RSS Feed Aggregator](https://en.wikipedia.org/wiki/Feed_aggregator) and reader written in [Go](https://go.dev/).

## Feature Summary

```
$ gator help
command: login :: Set user to named user >
        gator login user1
command: register :: Add new user to Gator >
        gator registered user2
command: reset :: WARNING: Wipe all users which results in cascade delete of all
        associated user records (feeds, follows, posts)
command: users :: List all
command: agg :: start a feed gathering loop to collect feed posts for followed
        feeds at specified interval >
        gator agg 15m
        m = minutes, 2h = hours
        To stop: ctrl-c
command: addfeed :: Add feed to availble feeds to follow. Start following the feed for
        the current user >
        gator addfeed "NASA News" https://cneos.jpl.nasa.gov/feed/news.xml
        * Any other user can elect to follow the feed using the "follow" command.,
        * Feed post will not be available unless you are running the "agg" command to gather feed data.
          This can be done in separate window since it keeps looping and gathering feed posts.
command: feeds :: List available feeds >
        gator feeds
command: follow :: Follow existing feed which was added previously by any user using
        the "addfeed" command >
        gator follow https://cneos.jpl.nasa.gov/feed/news.xml
        See the "feeds" command to list available feeds.
command: following :: List feeds already followed by currently logged in user >
        gator following
command: unfollow :: Unfollow a feed for the current user >
        gator unfollow https://cneos.jpl.nasa.gov/feed/news.xml
command: browse :: Browse most recently published feeds that were gathered by "agg" command >
        gator browse 20
```

## Install

### Install dependencies

#### PostgreSQL -- version 15 or later.

You may already have PostgreSQL installed as a dependency of another piece of software.

Verify with: `psql --version`

If the command is not found, download and install [Install PostgreSQL](https://www.postgresql.org/download/)

On most Linux based operating systems PostgreSQL will get installed under user
postgres with a default password of postgres. 
Any databases you create will also have a password with postgres as the
password. Unless you are playing in a sandboxed environment like a VM you do
not plan to keep you should change the postgres OS user password:

`sudo passwd postgres`

Start the PostgreSQL service if it is not running as software install does not
guarantee service is enabled and running.

Check with:

`ps -ef | \grep sql | grep -v grep`

If no output appears you will need to start the service. Note: to ensure it
starts again following a reboot you will want to enable the service.
These are two independent settings allowing you to perform one-time start/stop
versus always start the service whenever the OS is started.

Depending on your OS, these commands vary but, here is an example which should
work for most Linux distributions which use init.d or systemd to control
interfaces with some allowing both styles of commands to control and examine
services:
Init.d: `sudo service postgresql status`
systemd: `sudo systemctl status postgresql@18-main.service`

#### [Install Go Toolchain](https://go.dev/dl/)

## Configure Database (DB) and Gator

With PostgreSQL server running you will need to create a DB for Gator and
change the default password for DB which is also postgres like the
postgres OS user.

1. Access the PostgreSQL service using the basic psql client:
`sudo -u postgres psql`
2. From the "postgres=#" prompt:
`CREATE DATABASE gator;`
3. Connect to your new gator DB:
`\c gator`
4. Change the password for default user of your gator DB:
`ALTER USER postgres PASSWORD 'somethingsecret';`
5. You are done with the DB setup so you can:
`exit`


## Install Gator

1. Install the executable:
`go instal TODO`
2. TODO: TBD on how to bootstrap the DB !!!!
3. Create minimal Gator config:
`echo '{"db_url":"postgres://postgres:postgres@localhost:5432/gator?sslmode=disable"}' > ~/.gatorconfig.json`

Gator should now be ready to run:
`gator`

# Quickstart

1. `gator register $USER`
2. `gator users`
3. `gator addfeed "Tenable Security Advisories" https://www.tenable.com/security/feed`
4. Open a separate terminal window and start collecting feed posts for all
   available feeds. To start, use a small value so you can start browsing
   results:
`gator agg 1m`
5. After a few minutes, switch back to your first terminal and browse latest
   feed posts:
`gator browse 15`

# Environment variables

To control default logging levels where WARN is the default, use LOG_LEVEL like:
 `LOG_LEVEL=info gator`

Levels are those supported by the [Go slog package](https://pkg.go.dev/log/slog#Level)

On Windows in cmd.exe terminal shell:

`set LOG_LEVEL=info`

`gator`

To unset:

`set LOG_LEVEL=`
