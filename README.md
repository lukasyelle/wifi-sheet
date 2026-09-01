# wifi-sheet

Printable US Letter WiFi join sheets for macOS.

The CLI lists the WiFi networks saved on this Mac, lets you pick one or more, and writes a one-page PDF for each. The sheet has a PocketID-style mountain background, the network name, and a large QR code. Scanning the QR adds the network and connects. The password stays in the QR only. It is not printed on the page.

## Run

```bash
go run .
```

A filterable list of saved networks appears. Type `/` to search, space to toggle networks, and enter to generate. macOS will ask for Keychain access so the script can read the saved WiFi password. If the password is not in Keychain, the CLI asks you to type it. Each finished PDF opens in Preview.

Write sheets to another folder:

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
- Centered card with the SSID as the title
- The line `Scan to join this WiFi`
- A high-contrast QR code using the standard `WIFI:` payload that iPhone and Android understand

Print at 100% scale. Do not use "fit to page" if your printer already has unprintable margins. The card sits inside a safe area so the QR still scans.
