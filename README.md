# taskmail

`taskmail` allows you to generate an email report for [taskwarrior](https://taskwarrior.org/).

It expects the `task export` format. You can use the usual query and filtering features of taskwarrior to narrow down your list, and then export it to taskmail.

Setting up `taskmail` is easy.

    go get github.com/RaghavSood/taskmail

To be able to send emails, an SMTP config is required in `~/.taskmail/taskmail.yml`

A sample config looks like

    ---
    host: smtp.example.com
    port: 587
    password: hunter2
    to: example@example.com
    from: notexample@notexample.com

Invoking taskmail is as easy as:

    task export | taskmail

You can apply filters and queries to task too:

    task project:Home export | taskmail