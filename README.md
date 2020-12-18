# read-oots-cli

Giant in the Playground Games - Order of the Stick terminal comic reader.  This is
one of those tools that is really just written for myself.  I'm in a terminal window
every day, all day, so having the latest comic display when I open a new window is
a good way to keep up.  I had falled behind by more than a thousand strips.

Please visit [Order of the Stick](https://www.giantitp.com/comics/oots.html) comic strip
and if you enjoy the content please consider helping keep Rich free to entertain us.

[Patreon Subscription to OotS](https://www.patreon.com/oots)


## Dependencies

There are a few dependencies.  [mdcat](https://github.com/lunaryorn/mdcat) is used to display the comic in the
terminal window.  Review that page for supported terminal applications.

## Download and build the source

```bash
gh repo clone maahsome/read-oots-cli
cd read-oots-cli
go build
```

## Run the application to display

The applicatoin creates a tracking file `~/.config/read-oots-cli/oots` that contains the last
comic number displayed when using the `--next` and `--previous`, and does not when using the
`--number` flag, unless you also specify `--set-current`.  All these options can be reviewed
by running `read-oots-cli --help`.

```bash
./read-oots-cli
./read-oots-cli --next
./read-oots-cli --previous
./read-oots-cli --number 439
./read-oots-cli --number 1210 --set-current
```

## Copy the application to the path

```bash
sudo cp read-oots-cli /usr/local/bin/read-oots-cli
# optionally add an alias
alias oots=read-oots-cli
```

## Run on each login/open terminal window

I add the following line to my .zshrc (.bashrc) to have it check for a new comic and display it.

`/usr/local/bin/read-oots-cli --next`
