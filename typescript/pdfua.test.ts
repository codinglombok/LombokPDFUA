import { describe, it, expect } from 'vitest';
import PDFUA from './pdfua';

function bytes(s: string): Uint8Array {
  return new Uint8Array(Array.from(s).map((c) => c.charCodeAt(0)));
}

describe('PDFUA', () => {
  it('flags an invalid PDF as an error', () => {
    const pdfua = new PDFUA();
    pdfua.validate(bytes('not a pdf at all'));
    expect(pdfua.isCompliant()).toBe(false);
    const errors = pdfua.getErrors();
    expect(errors.length).toBe(1);
    expect(errors[0].rule).toBe('tagged_pdf_required');
  });

  it('flags an untagged PDF', () => {
    const pdfua = new PDFUA();
    pdfua.validate(bytes('%PDF-1.7\n%%EOF'));
    const rules = pdfua.getErrors().map((e) => e.rule);
    expect(rules).toContain('tagged_pdf_required');
  });

  it('does not flag a tagged PDF for structure', () => {
    const pdfua = new PDFUA();
    pdfua.validate(bytes('%PDF-1.7\n/StructTreeRoot 5 0 R\n%%EOF'));
    const rules = pdfua.getErrors().map((e) => e.rule);
    expect(rules).not.toContain('tagged_pdf_required');
  });

  it('flags a missing title as a warning', () => {
    const pdfua = new PDFUA();
    pdfua.validate(bytes('%PDF-1.7\n/StructTreeRoot 5 0 R\n%%EOF'));
    const rules = pdfua.getWarnings().map((w) => w.rule);
    expect(rules).toContain('document_title');
  });

  it('does not flag a present title', () => {
    const pdfua = new PDFUA();
    pdfua.validate(bytes('%PDF-1.7\n/StructTreeRoot 5 0 R\n/Title (My Doc)\n%%EOF'));
    const rules = pdfua.getWarnings().map((w) => w.rule);
    expect(rules).not.toContain('document_title');
  });

  it('flags a missing language as a warning', () => {
    const pdfua = new PDFUA();
    pdfua.validate(bytes('%PDF-1.7\n/StructTreeRoot 5 0 R\n%%EOF'));
    const rules = pdfua.getWarnings().map((w) => w.rule);
    expect(rules).toContain('language_specification');
  });

  it('flags an image without alt text as an error', () => {
    const pdfua = new PDFUA();
    pdfua.validate(bytes('%PDF-1.7\n/StructTreeRoot 5 0 R\n/Image <</Width 100>>\n%%EOF'));
    const rules = pdfua.getErrors().map((e) => e.rule);
    expect(rules).toContain('alt_text_for_images');
  });

  it('does not flag an image with alt text', () => {
    const pdfua = new PDFUA();
    pdfua.validate(bytes('%PDF-1.7\n/StructTreeRoot 5 0 R\n/Image <</Alt (A description)>>\n%%EOF'));
    const rules = pdfua.getErrors().map((e) => e.rule);
    expect(rules).not.toContain('alt_text_for_images');

    const images = pdfua.getImages();
    expect(images.length).toBe(1);
    expect(images[0].altText).toBe('A description');
    expect(images[0].hasAltText).toBe(true);
  });

  it('does not flag a decorative image without alt text', () => {
    const pdfua = new PDFUA();
    pdfua.validate(bytes('%PDF-1.7\n/StructTreeRoot 5 0 R\n/Image <</Artifact true>>\n%%EOF'));
    const rules = pdfua.getErrors().map((e) => e.rule);
    expect(rules).not.toContain('alt_text_for_images');
  });

  it('flags a form field without a label', () => {
    const pdfua = new PDFUA();
    pdfua.validate(bytes('%PDF-1.7\n/StructTreeRoot 5 0 R\n/FT /Tx /T (username)\n%%EOF'));
    const rules = pdfua.getErrors().map((e) => e.rule);
    expect(rules).toContain('form_labels');

    const fields = pdfua.getFormFields();
    expect(fields.length).toBe(1);
    expect(fields[0].name).toBe('username');
    expect(fields[0].type).toBe('Tx');
  });

  it('flags JavaScript as info-level', () => {
    const pdfua = new PDFUA();
    pdfua.validate(bytes('%PDF-1.7\n/StructTreeRoot 5 0 R\n/JavaScript (app.alert(1))\n%%EOF'));
    const rules = pdfua.getInfo().map((i) => i.rule);
    expect(rules).toContain('no_javascript');
  });

  it('is fully compliant for a well-formed document', () => {
    const pdfua = new PDFUA();
    pdfua.validate(
      bytes('%PDF-1.7\n/StructTreeRoot 5 0 R\n/Title (My Doc)\n/Lang (en-US)\n/H1\n%%EOF')
    );
    expect(pdfua.isCompliant()).toBe(true);
    expect(pdfua.isWCAGAACompliant()).toBe(true);
    expect(pdfua.getErrors().length).toBe(0);
  });

  it('returns correct issue counts', () => {
    const pdfua = new PDFUA();
    pdfua.validate(bytes('%PDF-1.7\n%%EOF')); // untagged, no title, no lang
    const [errors, warnings] = pdfua.getIssueCounts();
    expect(errors).toBe(1); // untagged
    expect(warnings).toBeGreaterThanOrEqual(2); // title + lang (+ heading)
  });

  it('has a fluent interface that returns self', () => {
    const pdfua = new PDFUA();
    const result = pdfua.validate(bytes('%PDF-1.7\n%%EOF'));
    expect(result).toBe(pdfua);
    expect(result.getReport()).not.toBeNull();
  });

  it('hardcodes the file name to document.pdf (known limitation)', () => {
    const pdfua = new PDFUA();
    pdfua.validate(bytes('%PDF-1.7\n%%EOF'));
    expect(pdfua.getReport()!.fileName).toBe('document.pdf');
  });

  it('serializes to JSON', () => {
    const pdfua = new PDFUA();
    pdfua.validate(
      bytes('%PDF-1.7\n/StructTreeRoot 5 0 R\n/Title (My Doc)\n/Lang (en-US)\n/H1\n%%EOF')
    );
    const json = JSON.parse(pdfua.toJSON());
    expect(json.isCompliant).toBe(true);
  });

  it('returns safe defaults before validation', () => {
    const pdfua = new PDFUA();
    expect(pdfua.isCompliant()).toBe(false);
    expect(pdfua.isWCAGAACompliant()).toBe(false);
    expect(pdfua.getErrors()).toEqual([]);
    expect(pdfua.toJSON()).toBe('{}');
  });
});
