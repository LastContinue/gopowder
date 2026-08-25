# Gopher Powder Ping (Gopowder)

Small Go program for checking Nix Darwin packages for updates

- **Gopher** Go Mascot
- **Powder** Snow, snowflake is Nix's icon
- **Ping** checking

This is a derivative name from the [Rust version](https://github.com/LastContinue/platypus-powder-ping) of this project. I decided to shorten it to something a big easier to type when CD'ing, and not to tongue-twist myself talking about the two different applications.

It still references an animal mascot (the excellent Go Gopher), so the criteria for "good name" is met as far as I'm concerned.

<img width="355" height="562" alt="output" src="https://github.com/user-attachments/assets/130ec2c4-fe9b-4e25-8d74-b90efb97774c" />

## Background

I originally wanted to find a way to see if I had any updates in my [Nix Darwin](https://github.com/nix-darwin/nix-darwin) setup.
I have a _very_ simple (read: "naïve and lazy") Flake based setup that I use to manage my entire install (no profiles, no channels, no HomeManager). Works great for me!

It's trivially easy to just `nix flake update` and then `darwin-rebuild switch` get updates but... when do I have updates? 🤔

Nix has a few ways of doing something _close_ to my idea, but none of them really clicked with my pea-brain, so I decided to just start hacking around in Fish (my shell) and see what I could come up with.

Tons of searching, some asking LLMs "why is this broken?", and a few hours later, I came out with something that would give me a list of packages that had a different version in `unstable` than I had on my machine.

"Close Enough!" I said.

Then I decided I wanted the format to be tabular... okay that's doable with just `column` it turns
out... but what about... and maybe this thing... and wow this is kinda slow and hard to work on now...
and what if I want to run it every day at X time... 😅

So I decided I would write an actual program to run the Nix commands and then wrangle the data that came from them. This would hopefully make
it easier for tabular output, progress spinners, progress bars, and putting into a .plist so [Launchd](https://en.wikipedia.org/wiki/Launchd) could run it every day at noon and give me a notification if there were updates. Oh and instead of running `Nix Command for Each Package` sequentially, I could run them all in parallel (hopefully much faster) without losing my mind (yes "fork process" is a thing. "No" I don't want to use that)

"Yes, but what language?..." I thought to myself. My first though was **Go**, as its concurrency mechanisms are top-notch, however, the current meta is "Rewrite it in Rust".

So I did! 🦀🦀🦀

You can find it here, [Platypus Powder Ping](https://github.com/LastContinue/platypus-powder-ping)

However, this repo is my "Original Vision", written in **Go** and it was fun to do something in a language I'm more familiar with, and see the different approaches to concurrency, error handling, etc first hand (I may write a think-piece about this at a later date).

Noteworthy, the performance between the two apps is roughly the same. This shocked me a bit because the Rust hype is "fast" but since the bottle neck is waiting on Nix, both languages can spin up concurrency faster than Nix can execute them.

## Setup

There's a `flake.nix` but its purely for running a Go dev shell (the extent of my Flake knowledge so far).  
I use `direnv` on my machine, so doing `cd`ing to the dir and then doing `direnv allow` should make everything work. YMMV depending on how you use Nix and/or if you already have a Go dev env.

`make test`, `make run` or `make build` then running the binary should get you up and running.

Before you run the actual program (via `make run` or via the binary) you'll need a toml file at `~/.config/gopowder/config.toml` (You can specify the file's location with the `-c` flag, but you will need a `config.toml` _somewhere_).

You can just copy the `config.toml.example` to that location and cut off the `.example` part. I could probably automate the creation of that, but "maybe later".

There should be some basic help with `-h`.

### A Note on Notifications

This program can send notifications. I learned some lessons about this topic from [the Rust version](https://github.com/LastContinue/platypus-powder-ping) so I went straight to **macos-notifier** and **osascript** for notifications so I could use the same api in this app as I had in the Rust version.

[macos-notifier](https://github.com/pixelperfectat/macos-cli-notifier)  
Blog post by author [here](https://thecoder.io/blog/native-macos-notifications-from-the-command-line-without-terminal-notifier)

<img width="360" height="97" alt="macos-notifier" src="https://github.com/user-attachments/assets/5c29a571-bb3f-43d3-a8e9-9d54b0c4c95b" />

</br></br>

> That is lame and I don't want to download anything

No problem. If `macos-notifier` isn't on your path, it will notify with `osascript` which, IMHO, isn't that good looking, but "it'll work".

If you don't care about notifications, then you're looking after your mental health, and I say "good for you!".
