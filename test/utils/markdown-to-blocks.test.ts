import { expect } from 'chai'
import { markdownToBlocks } from '../../src/utils/markdown-to-blocks'

/**
 * Tests for Markdown to Blocks Utility
 *
 * Verifies secure markdown conversion supporting:
 * - Headings (H1, H2, H3)
 * - Paragraphs with rich text (bold, italic, code, links)
 * - Bulleted lists
 * - Numbered lists
 * - Code blocks with language
 * - Blockquotes
 * - Horizontal rules
 * - Edge cases and malformed markdown
 */

describe('markdown-to-blocks', () => {
  describe('heading conversion', () => {
    it('should convert H1 to heading_1 block', () => {
      const result = markdownToBlocks('# Hello World')

      expect(result).to.have.lengthOf(1)
      expect(result[0].type).to.equal('heading_1')
      expect(result[0].heading_1).to.exist
      expect(result[0].heading_1.rich_text).to.have.lengthOf(1)
      expect(result[0].heading_1.rich_text[0].text.content).to.equal('Hello World')
    })

    it('should convert H2 to heading_2 block', () => {
      const result = markdownToBlocks('## Hello World')

      expect(result).to.have.lengthOf(1)
      expect(result[0].type).to.equal('heading_2')
      expect(result[0].heading_2).to.exist
      expect(result[0].heading_2.rich_text).to.have.lengthOf(1)
      expect(result[0].heading_2.rich_text[0].text.content).to.equal('Hello World')
    })

    it('should convert H3 to heading_3 block', () => {
      const result = markdownToBlocks('### Hello World')

      expect(result).to.have.lengthOf(1)
      expect(result[0].type).to.equal('heading_3')
      expect(result[0].heading_3).to.exist
      expect(result[0].heading_3.rich_text).to.have.lengthOf(1)
      expect(result[0].heading_3.rich_text[0].text.content).to.equal('Hello World')
    })

    it('should not convert H4+ as headings (fallback to paragraph)', () => {
      const result = markdownToBlocks('#### Hello World')

      expect(result).to.have.lengthOf(1)
      expect(result[0].type).to.equal('paragraph')
    })

    it('should handle headings with rich text formatting', () => {
      const result = markdownToBlocks('# Hello **World**')

      expect(result).to.have.lengthOf(1)
      expect(result[0].type).to.equal('heading_1')
      expect(result[0].heading_1.rich_text).to.have.lengthOf(2)
      expect(result[0].heading_1.rich_text[0].text.content).to.equal('Hello ')
      expect(result[0].heading_1.rich_text[1].text.content).to.equal('World')
      expect(result[0].heading_1.rich_text[1].annotations.bold).to.be.true
    })
  })

  describe('paragraph conversion', () => {
    it('should convert plain text to paragraph block', () => {
      const result = markdownToBlocks('This is a paragraph')

      expect(result).to.have.lengthOf(1)
      expect(result[0].type).to.equal('paragraph')
      expect(result[0].paragraph).to.exist
      expect(result[0].paragraph.rich_text).to.have.lengthOf(1)
      expect(result[0].paragraph.rich_text[0].text.content).to.equal('This is a paragraph')
    })

    it('should handle bold text with **', () => {
      const result = markdownToBlocks('This is **bold** text')

      expect(result).to.have.lengthOf(1)
      expect(result[0].paragraph.rich_text).to.have.lengthOf(3)
      expect(result[0].paragraph.rich_text[0].text.content).to.equal('This is ')
      expect(result[0].paragraph.rich_text[1].text.content).to.equal('bold')
      expect(result[0].paragraph.rich_text[1].annotations.bold).to.be.true
      expect(result[0].paragraph.rich_text[2].text.content).to.equal(' text')
    })

    it('should handle italic text with *', () => {
      const result = markdownToBlocks('This is *italic* text')

      expect(result).to.have.lengthOf(1)
      expect(result[0].paragraph.rich_text).to.have.lengthOf(3)
      expect(result[0].paragraph.rich_text[0].text.content).to.equal('This is ')
      expect(result[0].paragraph.rich_text[1].text.content).to.equal('italic')
      expect(result[0].paragraph.rich_text[1].annotations.italic).to.be.true
      expect(result[0].paragraph.rich_text[2].text.content).to.equal(' text')
    })

    it('should handle italic text with _', () => {
      const result = markdownToBlocks('This is _italic_ text')

      expect(result).to.have.lengthOf(1)
      expect(result[0].paragraph.rich_text).to.have.lengthOf(3)
      expect(result[0].paragraph.rich_text[0].text.content).to.equal('This is ')
      expect(result[0].paragraph.rich_text[1].text.content).to.equal('italic')
      expect(result[0].paragraph.rich_text[1].annotations.italic).to.be.true
      expect(result[0].paragraph.rich_text[2].text.content).to.equal(' text')
    })

    it('should handle inline code with `', () => {
      const result = markdownToBlocks('This is `code` text')

      expect(result).to.have.lengthOf(1)
      expect(result[0].paragraph.rich_text).to.have.lengthOf(3)
      expect(result[0].paragraph.rich_text[0].text.content).to.equal('This is ')
      expect(result[0].paragraph.rich_text[1].text.content).to.equal('code')
      expect(result[0].paragraph.rich_text[1].annotations.code).to.be.true
      expect(result[0].paragraph.rich_text[2].text.content).to.equal(' text')
    })

    it('should handle links with [text](url) format', () => {
      const result = markdownToBlocks('Visit [GitHub](https://github.com) for more')

      expect(result).to.have.lengthOf(1)
      expect(result[0].paragraph.rich_text).to.have.lengthOf(3)
      expect(result[0].paragraph.rich_text[0].text.content).to.equal('Visit ')
      expect(result[0].paragraph.rich_text[1].text.content).to.equal('GitHub')
      expect(result[0].paragraph.rich_text[1].text.link.url).to.equal('https://github.com')
      expect(result[0].paragraph.rich_text[2].text.content).to.equal(' for more')
    })

    it('should handle multiple formatting types in one paragraph', () => {
      const result = markdownToBlocks('This has **bold**, *italic*, `code`, and [links](https://example.com)')

      expect(result).to.have.lengthOf(1)
      expect(result[0].paragraph.rich_text).to.have.lengthOf(9)
      // Verify bold
      const boldText = result[0].paragraph.rich_text.find((rt: any) => rt.annotations?.bold)
      expect(boldText).to.exist
      expect(boldText.text.content).to.equal('bold')
      // Verify italic
      const italicText = result[0].paragraph.rich_text.find((rt: any) => rt.annotations?.italic)
      expect(italicText).to.exist
      expect(italicText.text.content).to.equal('italic')
      // Verify code
      const codeText = result[0].paragraph.rich_text.find((rt: any) => rt.annotations?.code)
      expect(codeText).to.exist
      expect(codeText.text.content).to.equal('code')
      // Verify link
      const linkText = result[0].paragraph.rich_text.find((rt: any) => rt.text.link)
      expect(linkText).to.exist
      expect(linkText.text.content).to.equal('links')
      expect(linkText.text.link.url).to.equal('https://example.com')
    })
  })

  describe('bulleted list conversion', () => {
    it('should convert - prefix to bulleted_list_item', () => {
      const result = markdownToBlocks('- Item one')

      expect(result).to.have.lengthOf(1)
      expect(result[0].type).to.equal('bulleted_list_item')
      expect(result[0].bulleted_list_item).to.exist
      expect(result[0].bulleted_list_item.rich_text[0].text.content).to.equal('Item one')
    })

    it('should convert * prefix to bulleted_list_item', () => {
      const result = markdownToBlocks('* Item one')

      expect(result).to.have.lengthOf(1)
      expect(result[0].type).to.equal('bulleted_list_item')
      expect(result[0].bulleted_list_item).to.exist
      expect(result[0].bulleted_list_item.rich_text[0].text.content).to.equal('Item one')
    })

    it('should handle multiple bulleted list items', () => {
      const markdown = `- Item one
- Item two
- Item three`
      const result = markdownToBlocks(markdown)

      expect(result).to.have.lengthOf(3)
      expect(result[0].type).to.equal('bulleted_list_item')
      expect(result[1].type).to.equal('bulleted_list_item')
      expect(result[2].type).to.equal('bulleted_list_item')
      expect(result[0].bulleted_list_item.rich_text[0].text.content).to.equal('Item one')
      expect(result[1].bulleted_list_item.rich_text[0].text.content).to.equal('Item two')
      expect(result[2].bulleted_list_item.rich_text[0].text.content).to.equal('Item three')
    })

    it('should handle rich text in bulleted list items', () => {
      const result = markdownToBlocks('- Item with **bold** text')

      expect(result).to.have.lengthOf(1)
      expect(result[0].bulleted_list_item.rich_text).to.have.lengthOf(3)
      expect(result[0].bulleted_list_item.rich_text[1].text.content).to.equal('bold')
      expect(result[0].bulleted_list_item.rich_text[1].annotations.bold).to.be.true
    })
  })

  describe('numbered list conversion', () => {
    it('should convert 1. prefix to numbered_list_item', () => {
      const result = markdownToBlocks('1. First item')

      expect(result).to.have.lengthOf(1)
      expect(result[0].type).to.equal('numbered_list_item')
      expect(result[0].numbered_list_item).to.exist
      expect(result[0].numbered_list_item.rich_text[0].text.content).to.equal('First item')
    })

    it('should handle multiple numbered list items', () => {
      const markdown = `1. First item
2. Second item
3. Third item`
      const result = markdownToBlocks(markdown)

      expect(result).to.have.lengthOf(3)
      expect(result[0].type).to.equal('numbered_list_item')
      expect(result[1].type).to.equal('numbered_list_item')
      expect(result[2].type).to.equal('numbered_list_item')
      expect(result[0].numbered_list_item.rich_text[0].text.content).to.equal('First item')
      expect(result[1].numbered_list_item.rich_text[0].text.content).to.equal('Second item')
      expect(result[2].numbered_list_item.rich_text[0].text.content).to.equal('Third item')
    })

    it('should handle rich text in numbered list items', () => {
      const result = markdownToBlocks('1. Item with `code` text')

      expect(result).to.have.lengthOf(1)
      expect(result[0].numbered_list_item.rich_text).to.have.lengthOf(3)
      expect(result[0].numbered_list_item.rich_text[1].text.content).to.equal('code')
      expect(result[0].numbered_list_item.rich_text[1].annotations.code).to.be.true
    })
  })

  describe('code block conversion', () => {
    it('should convert code block with language', () => {
      const markdown = '```javascript\nconst x = 42;\n```'
      const result = markdownToBlocks(markdown)

      expect(result).to.have.lengthOf(1)
      expect(result[0].type).to.equal('code')
      expect(result[0].code).to.exist
      expect(result[0].code.language).to.equal('javascript')
      expect(result[0].code.rich_text[0].text.content).to.equal('const x = 42;')
    })

    it('should convert code block without language (plain text)', () => {
      const markdown = '```\nconst x = 42;\n```'
      const result = markdownToBlocks(markdown)

      expect(result).to.have.lengthOf(1)
      expect(result[0].type).to.equal('code')
      expect(result[0].code.language).to.equal('plain text')
      expect(result[0].code.rich_text[0].text.content).to.equal('const x = 42;')
    })

    it('should handle multi-line code blocks', () => {
      const markdown = '```python\ndef hello():\n    print("Hello")\n    return True\n```'
      const result = markdownToBlocks(markdown)

      expect(result).to.have.lengthOf(1)
      expect(result[0].type).to.equal('code')
      expect(result[0].code.language).to.equal('python')
      expect(result[0].code.rich_text[0].text.content).to.include('def hello()')
      expect(result[0].code.rich_text[0].text.content).to.include('print("Hello")')
      expect(result[0].code.rich_text[0].text.content).to.include('return True')
    })

    it('should handle empty code blocks', () => {
      const markdown = '```\n```'
      const result = markdownToBlocks(markdown)

      expect(result).to.have.lengthOf(1)
      expect(result[0].type).to.equal('code')
      expect(result[0].code.rich_text[0].text.content).to.equal('')
    })
  })

  describe('blockquote conversion', () => {
    it('should convert > prefix to quote block', () => {
      const result = markdownToBlocks('> This is a quote')

      expect(result).to.have.lengthOf(1)
      expect(result[0].type).to.equal('quote')
      expect(result[0].quote).to.exist
      expect(result[0].quote.rich_text[0].text.content).to.equal('This is a quote')
    })

    it('should handle rich text in quotes', () => {
      const result = markdownToBlocks('> Quote with **bold** text')

      expect(result).to.have.lengthOf(1)
      expect(result[0].quote.rich_text).to.have.lengthOf(3)
      expect(result[0].quote.rich_text[1].text.content).to.equal('bold')
      expect(result[0].quote.rich_text[1].annotations.bold).to.be.true
    })

    it('should handle multiple quote lines as separate blocks', () => {
      const markdown = `> First quote
> Second quote`
      const result = markdownToBlocks(markdown)

      expect(result).to.have.lengthOf(2)
      expect(result[0].type).to.equal('quote')
      expect(result[1].type).to.equal('quote')
      expect(result[0].quote.rich_text[0].text.content).to.equal('First quote')
      expect(result[1].quote.rich_text[0].text.content).to.equal('Second quote')
    })
  })

  describe('horizontal rule conversion', () => {
    it('should convert --- to divider block', () => {
      const result = markdownToBlocks('---')

      expect(result).to.have.lengthOf(1)
      expect(result[0].type).to.equal('divider')
      expect(result[0].divider).to.deep.equal({})
    })

    it('should convert *** to divider block', () => {
      const result = markdownToBlocks('***')

      expect(result).to.have.lengthOf(1)
      expect(result[0].type).to.equal('divider')
      expect(result[0].divider).to.deep.equal({})
    })

    it('should convert ___ to divider block', () => {
      const result = markdownToBlocks('___')

      expect(result).to.have.lengthOf(1)
      expect(result[0].type).to.equal('divider')
      expect(result[0].divider).to.deep.equal({})
    })

    it('should handle longer sequences of rule characters', () => {
      const result = markdownToBlocks('-----')

      expect(result).to.have.lengthOf(1)
      expect(result[0].type).to.equal('divider')
    })
  })

  describe('edge cases', () => {
    it('should handle empty input', () => {
      const result = markdownToBlocks('')

      expect(result).to.be.an('array')
      expect(result).to.have.lengthOf(0)
    })

    it('should skip empty lines', () => {
      const markdown = `Line one

Line two`
      const result = markdownToBlocks(markdown)

      expect(result).to.have.lengthOf(2)
      expect(result[0].paragraph.rich_text[0].text.content).to.equal('Line one')
      expect(result[1].paragraph.rich_text[0].text.content).to.equal('Line two')
    })

    it('should handle unclosed formatting (treats as plain text)', () => {
      const result = markdownToBlocks('This has **unclosed bold')

      expect(result).to.have.lengthOf(1)
      // Should not crash, handles gracefully
      expect(result[0].paragraph.rich_text).to.exist
    })

    it('should handle unclosed link (treats as plain text)', () => {
      const result = markdownToBlocks('This has [unclosed link')

      expect(result).to.have.lengthOf(1)
      // Should not crash, handles gracefully
      expect(result[0].paragraph.rich_text).to.exist
    })

    it('should handle incomplete link syntax', () => {
      const result = markdownToBlocks('This has [text] without url')

      expect(result).to.have.lengthOf(1)
      // Should treat as plain text
      expect(result[0].paragraph.rich_text[0].text.content).to.include('[text]')
    })

    it('should handle nested formatting correctly', () => {
      const result = markdownToBlocks('This has **bold with `code`** inside')

      expect(result).to.have.lengthOf(1)
      // Should handle both formats
      expect(result[0].paragraph.rich_text).to.exist
      const hasBold = result[0].paragraph.rich_text.some((rt: any) => rt.annotations?.bold)
      const hasCode = result[0].paragraph.rich_text.some((rt: any) => rt.annotations?.code)
      expect(hasBold).to.be.true
      expect(hasCode).to.be.true
    })

    it('should add object and type to all blocks', () => {
      const result = markdownToBlocks('# Heading\nParagraph\n- List')

      expect(result).to.have.lengthOf(3)
      result.forEach(block => {
        expect(block.object).to.equal('block')
        expect(block.type).to.be.a('string')
      })
    })

    it('should handle mixed content types', () => {
      const markdown = `# Heading

Paragraph with **bold**

- List item 1
- List item 2

1. Numbered item

> Quote

\`\`\`javascript
code
\`\`\`

---`

      const result = markdownToBlocks(markdown)

      expect(result).to.have.lengthOf(9)
      expect(result[0].type).to.equal('heading_1')
      expect(result[1].type).to.equal('paragraph')
      expect(result[2].type).to.equal('bulleted_list_item')
      expect(result[3].type).to.equal('bulleted_list_item')
      expect(result[4].type).to.equal('numbered_list_item')
      expect(result[5].type).to.equal('quote')
      expect(result[6].type).to.equal('code')
      expect(result[7].type).to.equal('divider')
    })
  })

  describe('rich text parsing', () => {
    it('should handle text with no formatting', () => {
      const result = markdownToBlocks('Plain text')

      expect(result[0].paragraph.rich_text).to.have.lengthOf(1)
      expect(result[0].paragraph.rich_text[0].type).to.equal('text')
      expect(result[0].paragraph.rich_text[0].text.content).to.equal('Plain text')
    })

    it('should handle empty text gracefully', () => {
      const result = markdownToBlocks('> ')

      expect(result).to.have.lengthOf(1)
      expect(result[0].quote.rich_text).to.have.lengthOf(1)
      expect(result[0].quote.rich_text[0].text.content).to.equal('')
    })

    it('should preserve whitespace in text content', () => {
      const result = markdownToBlocks('Text   with   spaces')

      expect(result[0].paragraph.rich_text[0].text.content).to.equal('Text   with   spaces')
    })

    it('should handle special characters in text', () => {
      const result = markdownToBlocks('Text with <html> & special chars!')

      expect(result[0].paragraph.rich_text[0].text.content).to.equal('Text with <html> & special chars!')
    })
  })
})
