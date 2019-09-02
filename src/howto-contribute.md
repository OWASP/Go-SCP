How to Contribute
=================

This project is based on GitHub and can be accessed by [clicking here][15].

Here are the basic of contributing to GitHub:

1. Fork and clone the project
2. Set up the project locally
3. Create an upstream remote and sync your local copy
4. Branch each set of work
5. Push the work to your own repository
6. Create a new pull request
7. Look out for any code feedback and respond accordingly

This book was built from ground-up in a "collaborative fashion", using a small
set of Open Source tools and technologies.

Collaboration relies on [Git][1] - a free and open source distributed version
control system and other tools around Git:
* [Gogs][2] - Go Git Service, a painless self-hosted Git Service, which
  provides a Github like user interface and workflow.
* [Git flow][3] - a collection of Git extensions to provide high-level
  repository operations for [Vincent Driessen's branching model][4];
* [Git Flow Hooks][5] - some useful hooks for git-flow (AVH Edition) by
  [Jaspern Brouwer][6].

The book sources are written on [Markdown format][7], taking advantage of
[gitbook-cli][8].

## Environment setup

If you want to contribute to this book, you should setup the following tools on
your system:

1. To install Git, please follow the [official instructions][9]
   according to your system's configuration;
2. Now that you have Git, you should [install Git Flow][10] and
   [Git Flow Hooks][11];
3. Last but not least, [setup GitBook CLI][12].

## How to start

Ok, now you're ready to contribute.

Fork the `go-webapp-scp` repo and then clone your own repository.

The next step is to enable Git Flow hooks; enter your local repository

```shell
$ cd go-webapp-scp
```

and run

```shell
$ git flow init
```

We're good to go with git flow default values.

In a nutshell, everytime you want to work on a section, you should start a
"feature":

```shell
$ git flow feature start my-new-section
```

To keep your work safe, don't forget to publish your feature:

```shell
$ git flow feature publish
```

Once you're ready to merge your work with others, you should go to the main
repository and open a [Pull Request][14] to the `develop` branch. Then, someone
will review your work, leave any comments, request changes and/or simply merge
it on branch `develop` of project's main repository.

As soon as this happens, you'll need to pull the `develop` branch to keep your
own `develop` branch updated with the upstream. The same way as on a release,
you should update your `master` branch.

When you find a typo or something that needs to be fixed, you should start a
"hotfix"

```shell
$ git flow hotfix start
```

This will apply your change on both `develop` and `master` branches.

As you can see, until now there were no commits to the `master` branch. Great!
This is reserved for `releases`. When the work is ready to become publicly
available, the project owner will do the release.

While in the development stage, you can live-preview your work.
To get Git Book tracking file changes and to live-preview your work, you just
need to run the following command on a shell session

```shell
$ npm run serve
```

The shell output will include a `localhost` URL where you can preview the book.

## How to Build

If you have `node` installed, you can run:

```
$ npm i && node_modules/.bin/gitbook install && npm run build
```

You can also build the book using an ephemeral Docker container:

```
$ docker-compose run -u node:node --rm build
```

[1]: https://git-scm.com
[2]: https://gogs.io
[3]: https://github.com/petervanderdoes/gitflow-avh
[4]: http://nvie.com/posts/a-successful-git-branching-model
[5]: https://github.com/jaspernbrouwer/git-flow-hooks
[6]: https://github.com/jaspernbrouwer
[7]: http://daringfireball.net/projects/markdown
[8]: https://github.com/GitbookIO/gitbook-cli
[9]: https://git-scm.com/downloads
[10]: https://github.com/petervanderdoes/gitflow-avh/wiki/Installation
[11]: https://github.com/jaspernbrouwer/git-flow-hooks#install
[12]: https://github.com/GitbookIO/gitbook-cli#how-to-install-it
[14]: http://help.github.com/articles/about-pull-requests
[15]: https://github.com/Checkmarx/Go-SCP
[16]: https://www.docker.com/
[17]: https://nodejs.org/en/
[18]: https://calibre-ebook.com/download
