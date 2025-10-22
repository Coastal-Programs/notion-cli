import { expect } from 'chai'
import {
  expandSimpleProperties,
  validateSimpleProperties,
  SimpleProperties,
} from '../../src/utils/property-expander'

describe('Property Expander', () => {
  // Mock database schema
  const mockSchema = {
    Name: {
      id: 'title',
      type: 'title',
      title: {},
    },
    Status: {
      id: 'status',
      type: 'select',
      select: {
        options: [
          { id: '1', name: 'Not Started', color: 'default' },
          { id: '2', name: 'In Progress', color: 'blue' },
          { id: '3', name: 'Done', color: 'green' },
        ],
      },
    },
    Priority: {
      id: 'priority',
      type: 'number',
      number: { format: 'number' },
    },
    'Due Date': {
      id: 'due',
      type: 'date',
      date: {},
    },
    Tags: {
      id: 'tags',
      type: 'multi_select',
      multi_select: {
        options: [
          { id: '1', name: 'urgent', color: 'red' },
          { id: '2', name: 'bug', color: 'orange' },
          { id: '3', name: 'feature', color: 'purple' },
        ],
      },
    },
    Completed: {
      id: 'completed',
      type: 'checkbox',
      checkbox: {},
    },
    Email: {
      id: 'email',
      type: 'email',
      email: {},
    },
    URL: {
      id: 'url',
      type: 'url',
      url: {},
    },
    Description: {
      id: 'desc',
      type: 'rich_text',
      rich_text: {},
    },
  }

  describe('expandSimpleProperties', () => {
    it('should expand title property', async () => {
      const simple: SimpleProperties = {
        Name: 'My Task',
      }

      const result = await expandSimpleProperties(simple, mockSchema)

      expect(result).to.deep.equal({
        Name: {
          title: [{ text: { content: 'My Task' } }],
        },
      })
    })

    it('should expand select property with case-insensitive matching', async () => {
      const simple: SimpleProperties = {
        Status: 'in progress', // lowercase
      }

      const result = await expandSimpleProperties(simple, mockSchema)

      expect(result).to.deep.equal({
        Status: {
          select: { name: 'In Progress' }, // Exact case from schema
        },
      })
    })

    it('should expand number property', async () => {
      const simple: SimpleProperties = {
        Priority: 5,
      }

      const result = await expandSimpleProperties(simple, mockSchema)

      expect(result).to.deep.equal({
        Priority: {
          number: 5,
        },
      })
    })

    it('should expand date property with ISO date', async () => {
      const simple: SimpleProperties = {
        'Due Date': '2025-12-31',
      }

      const result = await expandSimpleProperties(simple, mockSchema)

      expect(result).to.deep.equal({
        'Due Date': {
          date: { start: '2025-12-31' },
        },
      })
    })

    it('should expand date property with relative date (today)', async () => {
      const simple: SimpleProperties = {
        'Due Date': 'today',
      }

      const result = await expandSimpleProperties(simple, mockSchema)

      const today = new Date().toISOString().split('T')[0]
      expect(result['Due Date'].date.start).to.equal(today)
    })

    it('should expand date property with relative date (tomorrow)', async () => {
      const simple: SimpleProperties = {
        'Due Date': 'tomorrow',
      }

      const result = await expandSimpleProperties(simple, mockSchema)

      const tomorrow = new Date()
      tomorrow.setDate(tomorrow.getDate() + 1)
      const expectedDate = tomorrow.toISOString().split('T')[0]

      expect(result['Due Date'].date.start).to.equal(expectedDate)
    })

    it('should expand date property with relative date (+7 days)', async () => {
      const simple: SimpleProperties = {
        'Due Date': '+7 days',
      }

      const result = await expandSimpleProperties(simple, mockSchema)

      const sevenDaysFromNow = new Date()
      sevenDaysFromNow.setDate(sevenDaysFromNow.getDate() + 7)
      const expectedDate = sevenDaysFromNow.toISOString().split('T')[0]

      expect(result['Due Date'].date.start).to.equal(expectedDate)
    })

    it('should expand multi-select property', async () => {
      const simple: SimpleProperties = {
        Tags: ['urgent', 'bug'],
      }

      const result = await expandSimpleProperties(simple, mockSchema)

      expect(result).to.deep.equal({
        Tags: {
          multi_select: [{ name: 'urgent' }, { name: 'bug' }],
        },
      })
    })

    it('should expand checkbox property (boolean)', async () => {
      const simple: SimpleProperties = {
        Completed: true,
      }

      const result = await expandSimpleProperties(simple, mockSchema)

      expect(result).to.deep.equal({
        Completed: {
          checkbox: true,
        },
      })
    })

    it('should expand checkbox property (string)', async () => {
      const simple: SimpleProperties = {
        Completed: 'yes',
      }

      const result = await expandSimpleProperties(simple, mockSchema)

      expect(result).to.deep.equal({
        Completed: {
          checkbox: true,
        },
      })
    })

    it('should expand email property', async () => {
      const simple: SimpleProperties = {
        Email: 'test@example.com',
      }

      const result = await expandSimpleProperties(simple, mockSchema)

      expect(result).to.deep.equal({
        Email: {
          email: 'test@example.com',
        },
      })
    })

    it('should expand URL property', async () => {
      const simple: SimpleProperties = {
        URL: 'https://example.com',
      }

      const result = await expandSimpleProperties(simple, mockSchema)

      expect(result).to.deep.equal({
        URL: {
          url: 'https://example.com',
        },
      })
    })

    it('should expand rich_text property', async () => {
      const simple: SimpleProperties = {
        Description: 'This is a description',
      }

      const result = await expandSimpleProperties(simple, mockSchema)

      expect(result).to.deep.equal({
        Description: {
          rich_text: [{ text: { content: 'This is a description' } }],
        },
      })
    })

    it('should expand multiple properties at once', async () => {
      const simple: SimpleProperties = {
        Name: 'Complex Task',
        Status: 'Done',
        Priority: 10,
        'Due Date': '2025-12-31',
        Tags: ['urgent', 'feature'],
        Completed: true,
      }

      const result = await expandSimpleProperties(simple, mockSchema)

      expect(result).to.have.all.keys(['Name', 'Status', 'Priority', 'Due Date', 'Tags', 'Completed'])
      expect(result.Name.title[0].text.content).to.equal('Complex Task')
      expect(result.Status.select.name).to.equal('Done')
      expect(result.Priority.number).to.equal(10)
      expect(result.Completed.checkbox).to.equal(true)
    })

    it('should handle null values', async () => {
      const simple: SimpleProperties = {
        Description: null,
      }

      const result = await expandSimpleProperties(simple, mockSchema)

      expect(result).to.deep.equal({
        Description: null,
      })
    })

    it('should handle case-insensitive property names', async () => {
      const simple: SimpleProperties = {
        name: 'Task', // lowercase
        STATUS: 'Done', // uppercase
        'due date': '2025-12-31', // lowercase with space
      }

      const result = await expandSimpleProperties(simple, mockSchema)

      expect(result).to.have.all.keys(['Name', 'Status', 'Due Date'])
      expect(result.Name.title[0].text.content).to.equal('Task')
      expect(result.Status.select.name).to.equal('Done')
      expect(result['Due Date'].date.start).to.equal('2025-12-31')
    })

    it('should throw error for unknown property', async () => {
      const simple: SimpleProperties = {
        UnknownProp: 'value',
      }

      try {
        await expandSimpleProperties(simple, mockSchema)
        expect.fail('Should have thrown an error')
      } catch (error: any) {
        expect(error.message).to.include('Property "UnknownProp" not found')
        expect(error.message).to.include('Available properties:')
      }
    })

    it('should throw error for invalid select option', async () => {
      const simple: SimpleProperties = {
        Status: 'Invalid Status',
      }

      try {
        await expandSimpleProperties(simple, mockSchema)
        expect.fail('Should have thrown an error')
      } catch (error: any) {
        expect(error.message).to.include('Invalid select value')
        expect(error.message).to.include('Valid options:')
      }
    })

    it('should throw error for invalid multi-select option', async () => {
      const simple: SimpleProperties = {
        Tags: ['urgent', 'invalid-tag'],
      }

      try {
        await expandSimpleProperties(simple, mockSchema)
        expect.fail('Should have thrown an error')
      } catch (error: any) {
        expect(error.message).to.include('Invalid multi-select value')
        expect(error.message).to.include('invalid-tag')
      }
    })

    it('should throw error for invalid email', async () => {
      const simple: SimpleProperties = {
        Email: 'not-an-email',
      }

      try {
        await expandSimpleProperties(simple, mockSchema)
        expect.fail('Should have thrown an error')
      } catch (error: any) {
        expect(error.message).to.include('Invalid email')
      }
    })

    it('should throw error for invalid URL', async () => {
      const simple: SimpleProperties = {
        URL: 'not-a-url',
      }

      try {
        await expandSimpleProperties(simple, mockSchema)
        expect.fail('Should have thrown an error')
      } catch (error: any) {
        expect(error.message).to.include('Invalid URL')
      }
    })

    it('should throw error for invalid number', async () => {
      const simple: SimpleProperties = {
        Priority: 'not-a-number',
      }

      try {
        await expandSimpleProperties(simple, mockSchema)
        expect.fail('Should have thrown an error')
      } catch (error: any) {
        expect(error.message).to.include('Invalid number value')
      }
    })
  })

  describe('validateSimpleProperties', () => {
    it('should validate correct properties', () => {
      const simple: SimpleProperties = {
        Name: 'Task',
        Status: 'Done',
        Priority: 5,
      }

      const result = validateSimpleProperties(simple, mockSchema)

      expect(result.valid).to.be.true
      expect(result.errors).to.be.empty
    })

    it('should detect unknown properties', () => {
      const simple: SimpleProperties = {
        UnknownProp: 'value',
      }

      const result = validateSimpleProperties(simple, mockSchema)

      expect(result.valid).to.be.false
      expect(result.errors).to.have.lengthOf(1)
      expect(result.errors[0]).to.include('UnknownProp')
    })

    it('should detect invalid values', () => {
      const simple: SimpleProperties = {
        Email: 'invalid-email',
        URL: 'invalid-url',
      }

      const result = validateSimpleProperties(simple, mockSchema)

      expect(result.valid).to.be.false
      expect(result.errors.length).to.be.greaterThan(0)
    })

    it('should collect multiple errors', () => {
      const simple: SimpleProperties = {
        UnknownProp: 'value',
        Email: 'invalid',
        Status: 'Invalid Status',
      }

      const result = validateSimpleProperties(simple, mockSchema)

      expect(result.valid).to.be.false
      expect(result.errors.length).to.be.greaterThan(2)
    })
  })
})
