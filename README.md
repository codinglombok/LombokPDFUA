# LombokPDFUA

PDF/UA accessibility compliance library (ISO 14289).  
Part of the [Lombok Ecosystem](https://github.com/codinglombok).

## Features

- Tagged PDF structure (semantic tags)
- Alt text for images
- Form accessibility
- Navigation structure
- WCAG 2.1 AA validation
- Zero external dependencies
- Multi-language: TypeScript, Python, PHP, Go

## Installation

```bash
npm install lombokpdfua          # TypeScript/JavaScript
pip install lombokpdfua          # Python
composer require codinglombok/lombokpdfua  # PHP
go get github.com/codinglombok/lombokpdfua # Go
```

## Quick Start (TypeScript)

```typescript
import { PDFUA } from 'lombokpdfua'

const pdfua = await new PDFUA(pdfBuffer)
  .tagStructure()
  .addAltText({ images: true, headings: true })
  .validateAccessibility()
  .export()
```

## Use Cases

- Government/public documents (ADA compliance)
- Educational materials
- Accessible web PDFs
- Screen-reader friendly documents

## Project Structure

```
lombokpdfua/
├── typescript/     # TypeScript source + tests
├── python/         # Python port + tests
├── go/             # Go port + tests
├── php/            # PHP port + tests
├── package.json
├── LICENSE         # Apache 2.0
└── README.md
```

## License

Apache 2.0 — see [LICENSE](LICENSE)
