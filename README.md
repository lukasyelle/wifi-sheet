# wifi-sheet

Printable US Letter WiFi join sheets for macOS.

The CLI lists the WiFi networks saved on this Mac, lets you pick one or more, and writes a single one-page PDF. Every selected network appears on that page with its own QR code. The sheet has a PocketID-style mountain background, each network name, and high-contrast QR codes. Scanning a QR adds that network and connects. Passwords stay in the QR only. They are not printed on the page.

## Run

```bash
go run .
```

A filterable list of saved networks appears. Type `/` to search, space to toggle networks, and enter to generate. macOS will ask for Keychain access so the script can read the saved WiFi password. If the password is not in Keychain, the CLI asks you to type it. The finished PDF opens in Preview.

Write the sheet to another folder:

```bash
go run . -o ~/Desktop
```

Build a binary:

```bash
go build -o wifi-sheet .
./wifi-sheet -o ~/Desktop
```

## What you get

Each PDF is a single US Letter page:

- Full-page nature background
- One card per selected network with the SSID as the title
- The line `Scan to join this WiFi` (or a shorter `Scan to join` when several fit on the page)
- A high-contrast QR code per network using the standard `WIFI:` payload that iPhone and Android understand

One network keeps the large centered card. Two or more networks share the page in a grid so every QR stays scannable.

Print at 100% scale. Do not use "fit to page" if your printer already has unprintable margins. The cards sit inside a safe area so the QR codes still scan.
