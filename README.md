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

If the command is not found, use your preferred package manager to install
the software. If not packaged for your system you can
download and [Install PostgreSQL](https://www.postgresql.org/download/)
directly.

On some Linux based operating systems PostgreSQL will get installed under user
postgres with a default password of postgres.
Any databases you create will also have a password with postgres as the
password. Unless you are playing in a sand boxed environment like a VM you do
not plan to keep you should change the postgres OS user password:

`sudo passwd postgres`

On other Linux systems, sudo model eliminates the need to directly access the
postgres user or know its credentials.

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

If the service fails to start, it's because these things are not as simple as
often indicated and on a fresh install the installer chooses not to
set up a container directory where your databases are managed.

```
# On a fresh Manjaro XFCE VM
sudo pacman -S postgresql
sudo systemctl status postgresql.service
sudo systemctl enable postgresql.service
sudo systemctl status postgresql.service
sudo systemctl start postgresql.service
sudo systemctl status postgresql.service

# When service failed to start, status gives you a hint on where to find trouble-shooting information
sudo journalctl -xeu postgresql.service

# A captured message from the Postgres failed startup indicates missing data dir and how to create it:
su -l postgres -c "initdb --locale=C.UTF-8 --encoding=UTF8 -D '/var/lib/postgres/data'"

# However this is not the way for sudo based systems. When prompted, you will enter your privileged user password which will give you access to postgres user to run the command:
sudo -u postgres initdb --locale=C.UTF-8 --encoding=UTF8 -D '/var/lib/postgres/data'
sudo systemctl start postgresql.service && sudo systemctl status postgresql.service

# You should now see the DB service running.
```

#### [Install Go Toolchain](https://go.dev/dl/)

## Configure Database (DB) and Gator

With PostgreSQL server running you will need to create a DB for Gator.
Note: On some systems, the postgres DB user may have a default password of
postgres applied.
1. Access the PostgreSQL service using the basic psql client:
`sudo -u postgres psql`
2. From the "postgres=#" prompt:
`CREATE DATABASE gator;`
3. Connect to your new gator DB:
`\c gator`
4. To change or apply password for your gator DB:
`ALTER USER postgres PASSWORD 'somethingsecret';`
5. You are done with the DB setup so you can:
`exit`


## Install Gator

1. Install the executable:

   `go instal https://github.com/dragonbitestail/gator@latest`
2. Bootstrap the Gator DB schema:
   - Install Goose DB migration tool: `go install github.com/pressly/goose/v3/cmd/goose@latest`
   - Find cached Gator `sql` directory which was created when you installed Gator (1):

     `find ~/go -type d -name sql |grep gator`

     If you have several, cd to the most recent. E.g.:

     `cd /home/dbt/go/pkg/mod/github.com/dragonbitestail/gator@v0.0.0-20260925143502-a7f8580135b2/sql/schema`

   - Run the migrations with Goose using an appropriate connection string:

     `goose postgres "postgres://postgres:@localhost:5432/gator" up`

     If all went well you should see a success message.
3. Create minimal Gator configuration. Once, configured this is managed by
   Gator:

   ```
   echo '{"db_url":"postgres://postgres@localhost:5432/gator?sslmode=disable"}' > ~/.gatorconfig.json
   ```

Gator should now be ready to run:

`gator`

# Quick start

1. `gator register $USER`
2. `gator users`
3. `gator addfeed "NASA News" https://cneos.jpl.nasa.gov/feed/news.xml`
4. Open a separate terminal window and start collecting feed posts for all
   available feeds. To start, use a small value so you can start browsing
   results:

   `gator agg 1m`
5. After a few minutes, switch back to your first terminal and browse latest
   feed posts:

   `gator browse 15`
6. Remember to exit the aggregator when you are done testing. If you wish to
   run it regularly, use a more reasonable value like like every 2h in the case
   where may check feeds a few times a day:

   `gator agg 2h`

# Environment variables

To control default logging levels where WARN is the default, use LOG_LEVEL like:
 `LOG_LEVEL=info gator`

Levels are those supported by the [Go slog package](https://pkg.go.dev/log/slog#Level)
