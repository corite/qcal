# qcal

qcal is a quick calendar application for CalDAV servers written in Go. In
contrast to other tools it does not cache anything. It can fetch multiple
servers / calendars in parallel which makes it quite fast.

Its main purpose is displaying calendar data. Nevertheless it supports basic
creation and editing of entries.

## Features

- Condensed quick overview of appointments
- Parallel fetching of multiple calendars
- Viewing of iCal subscriptions
- Easy to use filters
- Create, modify and delete appointments
- Import ICS files
- Display ICS files
- Easy setup


## Installation / Configuration

- Have Go installed
- make && sudo make install (for MacOS: make darwin)
- copy config-sample.json to ~/.config/qcal/config.json and modify accordingly

### Arch AUR package

- Here: [AUR](https://aur.archlinux.org/packages/qcal)
- Copy config-sample.json from /usr/share/qcal/ to ~/.config/qcal/config.json and modify accordingly

## Configuration

- For additional calendars just add a comma and new calendar credentials in
  curly brackets.
- Omit Username and Password if you add a readonly iCal source


## Usage

Common options:

    qcal -h

### Displaying appointments

This simply displays all appointments for the next x days (configured in config.json):

    qcal

This shows all appointments for today:

    qcal -t

This only shows appointments from calendar 0 for the next seven days:

    qcal -c 0 -7

This shows all appointments from 01.10.2021, 00:00h to 31.10.2021, 23:59:59
(Note: This is in UTC!):

    qcal -s 20211001T000000 -e 20211031T235959

This displays all avaliable calendars with their numbers and colors:

    qcal -l

### Add new appointment

Even though the abillity to create new appointments is limited, it is easy to
create simple appointment types.

This creates an appointment on 01.12.2021 from 15:00h to 17:00h with the
summary of "Tea Time":

    qcal -n "20211201 1500 1700 Tea Time"

This creates a whole day appointment with a yearly recurrence in your second
calendar (first is 0):

    qcal -c 1 -n "20211114 Anne's Birthday" -r y

This creates a multiple day appointment:

    qcal -n "20210801 20210810 Holiday in Thailand"

### Edit an appointment

This shows the next 7 days of appointments from calendar 3 with filenames
("foobarxyz.ics"):

    qcal -c 2 -7 -f 

This edits the selected iCAL object in your $EDITOR (i.e. vim). When you
save-quit the modified object is automatically uploaded:

    qcal -c 2 -edit foobarxyz.ics


## Integrations

### neomutt / other cli mail tools

You can view received appointments in neomutt with qcal! Put this in your
mailcap (usually in .config/neomutt):

    text/calendar; qcal -p; copiousoutput

### Crontab 

You can get reminders of your appointments 15 mins in advance with this one
liner:

    EVENT=$(qcal -cron 15); [[ $EVENT ]] && notify-send "Next Appointment:" "\n$EVENT"


## About

Questions? Ideas? File bugs and TODOs through the [issue
tracker](https://todo.sr.ht/~psic4t/qcal) or send an email to
[~psic4t/qcal@todo.sr.ht](mailto:~psic4t/qcal@todo.sr.ht)
