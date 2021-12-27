# qcal

qcal is a quick calendar application for CalDAV servers written in Go. In
contrast to other tools it does not cache anything. It can fetch multiple
servers / calendars in parallel which makes it quite fast.

Its main purpose is displaying calendar data. Nevertheless it supports basic
creation and editing of entries.

## Features

- condensed quick overview of appointments
- parallel fetching of multiple calendars
- easy to use filters
- create, modify and delete appointments
- import ICS files
- display ICS files
- easy setup


## Installation

- Have Go installed
- make && sudo make install
- copy config-sample.json to ~/.config/qcal/config.json and modify accordingly


## Usage

- qcal -h for common options

### Add new appointment

Even though the abillity to create new appointments is limited, it is easy to create simple appointment types:

    qcal -n "20211201 1500 1700 Tea Time"

This crates an appointment on 01.12.2021 from 15:00h to 17:00h with the summary of "Tea Time"

    qcal -n "20211114 Anne's Birthday" -r y

This creates a whole day appointment with a yearly recurrence

    qcal -n "20210801 20210810 Holiday in Thailand"

This creates a multiple day appointment


## neomutt / other cli mail tools

You can view received appointments in neomutt with qcal! Put this in your
mailcap (usually in .config/neomutt):

    text/calendar; qcal -p; copiousoutput


## Crontab 

You can get reminders of your appointments 15 mins in advance with this one liner:

    EVENT=$(qcal -cron 15); [[ $EVENT ]] && notify-send "Next Appointment:" "\n$EVENT"

