# remindme - quick n stupid desktop reminders and simple pomodoro timer

## Install

Arch: `yay -S remindme-bin`

## Usage

- `remindme in <duration><m/h> <message>`, example: `remindme in 10m This is a message`
- `remindme at <time> <message>`, example: `remindme at 15:32 This is a message`
- `remindme p <start/stop>`
- `remindme notify <sound>`, example: `remindme notify message`
- `remindme --watch`, starts notifier

Adding a "!" suffix to the message will create an urgent notification instead of a normal one.

`<sound>` is a sound file within `/usr/share/sounds/freedesktop/stereo/`.

## Pomodoro

- 25 minutes work
- 5 minute break
- 20 minutes break after 4 working units

## quick n dirty

dont look at the code. it works.
