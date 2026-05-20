# 📄 Paperless-NGX Document Overview

A Go command-line tool that fetches documents from a [Paperless-NGX](https://docs.paperless-ngx.com/) instance,
groups them by year and generates a nicely formatted PDF report per correspondent.

## ✨ Features

- 📋 Fetches all documents for configured correspondents via Paperless-NGX REST API
- 🗂️ Groups documents by year with subtotals
- 💰 Parses and sums CHF amounts from custom fields
- 📄 Generates a PDF report per correspondent including:
  - Clickable document ID links back to Paperless-NGX
  - Year sections with totals
  - Grand total
  - Missing amount highlighting in red
  - Alternating row colors
- 🖥️ Prints a formatted summary table to the terminal
- 🔍 Optional filter to show only invoices (`InvoiceOnly`)

## 📁 Project Structure

