# Check-break
[![BCH compliance](https://bettercodehub.com/edge/badge/Prytoegrian/check-break?branch=master)](https://bettercodehub.com/)
[![Codacy Badge](https://api.codacy.com/project/badge/Grade/6f9a3e0c93ce4fb28c778019ad083179)](https://www.codacy.com/app/prytoegrian/check-break?utm_source=github.com&amp;utm_medium=referral&amp;utm_content=Prytoegrian/check-break&amp;utm_campaign=Badge_Grade)

`check-break` helps you to discover compatibility breaks in your code, and improve decisions-making to determine if a new major version is required. In few words, if you follow *semver* (or try to stick to it), you must use `check-break` ;-)

`check-break` itself is following `semver` and monitors its own changes. For now, it's under heavy development, use it at your own risks, compatibility breaks could happen at every moment.

## What is a compatibility break ?
Basically, following the semver definition :  
> A change is incompatible if it removes a possibility for the consumer in the public API.

All starts with a clear definition of **public API** in your context. Once done, a compatibility break occurs on all **public API** functions each time :
- an argument is removed
- an argument is added (*without a default value*)
- a default argument is removed
- a return type is removed
- a return type is added (*without a default value*)
- type of any input / output / exception / assertion is changed and is incompatible with the former one (**1**)

**1.** In other words, if you're comfortable with [Liskov principle](https://en.wikipedia.org/wiki/Liskov_substitution_principle), you might have heard :
> Be contravariant in yours preconditions, be covariant with yours postconditions.

Thus, a compatibility break happens when you're covariant in yours preconditions or contravariant in yours postconditions.

Since `check-break` can't guess your public API (yet), it shows you all changes on public functions. It's up to you to determine if :
- this change really is a break,
- this change is in the public API scope.

## First launch
(*Assuming there's a golang env on your system*)
```sh
make install
```
That's it.

## Usage
This tool is based upon `git`, and particularly on diff between two points. Thus, the syntax is as follows :
```sh
$ check-break -s starting_point -e ending_point [-p path_to_git_repository] [-c path_to_config]
```

**Note:** All unsupported files are also reported as such, in order not to give a feeling of false negative.

## Langages supported

Obviously, I started with langages I use in a daily-basis :
- Go
- Java
- Javascript
- PHP
- sh

Feel free to participate to add yours, correct bugs, improve software design, etc. `check-break` is under [GPL3](LICENCE).

Please remember that this tool may be incomplete, it doesn't replace the human judgment.
