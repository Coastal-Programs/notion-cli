"use strict";
/**
 * Property Example Generator for Notion API
 *
 * Generates copy-pastable property payload examples based on database schema.
 * Helps AI agents understand the correct format for create/update operations.
 */
Object.defineProperty(exports, "__esModule", { value: true });
exports.generatePropertyExamples = generatePropertyExamples;
exports.formatExamplesForConsole = formatExamplesForConsole;
exports.groupExamplesByWritability = groupExamplesByWritability;
/**
 * Generate property examples for all properties in a data source schema
 *
 * @param properties - Properties object from GetDataSourceResponse
 * @returns Array of property examples
 */
function generatePropertyExamples(properties) {
    const examples = [];
    for (const [propName, propDef] of Object.entries(properties)) {
        const example = generateExampleForType(propName, propDef);
        if (example) {
            examples.push(example);
        }
    }
    return examples;
}
/**
 * Generate example for a single property based on its type
 *
 * @param name - Property name
 * @param propDef - Property definition from Notion API
 * @returns Property example or null if unsupported
 */
function generateExampleForType(name, propDef) {
    var _a, _b, _c, _d, _e, _f, _g, _h, _j;
    if (!propDef || !propDef.type) {
        return null;
    }
    const type = propDef.type;
    switch (type) {
        case 'title':
            return {
                property_name: name,
                property_type: 'title',
                simple_value: 'My Page Title',
                notion_payload: {
                    [name]: {
                        title: [{ text: { content: 'My Page Title' } }]
                    }
                },
                description: 'Main title of the page (required for new pages)'
            };
        case 'rich_text':
            return {
                property_name: name,
                property_type: 'rich_text',
                simple_value: 'Some text content',
                notion_payload: {
                    [name]: {
                        rich_text: [{ text: { content: 'Some text content' } }]
                    }
                },
                description: 'Multi-line text with optional formatting'
            };
        case 'number': {
            const numberFormat = ((_a = propDef.number) === null || _a === void 0 ? void 0 : _a.format) || 'number';
            return {
                property_name: name,
                property_type: 'number',
                simple_value: 42,
                notion_payload: {
                    [name]: { number: 42 }
                },
                description: `Numeric value (format: ${numberFormat})`
            };
        }
        case 'checkbox':
            return {
                property_name: name,
                property_type: 'checkbox',
                simple_value: true,
                notion_payload: {
                    [name]: { checkbox: true }
                },
                description: 'Boolean true/false value'
            };
        case 'select': {
            const selectOptions = ((_b = propDef.select) === null || _b === void 0 ? void 0 : _b.options) || [];
            const firstOption = ((_c = selectOptions[0]) === null || _c === void 0 ? void 0 : _c.name) || 'Option Name';
            const selectOptionsList = selectOptions.map((o) => o.name).join(', ');
            return {
                property_name: name,
                property_type: 'select',
                simple_value: firstOption,
                notion_payload: {
                    [name]: { select: { name: firstOption } }
                },
                description: selectOptions.length > 0
                    ? `Single selection from: ${selectOptionsList}`
                    : 'Single selection (no options defined yet)'
            };
        }
        case 'multi_select': {
            const multiOptions = ((_d = propDef.multi_select) === null || _d === void 0 ? void 0 : _d.options) || [];
            const exampleOptions = multiOptions.slice(0, 2).map((o) => o.name);
            const multiOptionsList = multiOptions.map((o) => o.name).join(', ');
            return {
                property_name: name,
                property_type: 'multi_select',
                simple_value: exampleOptions,
                notion_payload: {
                    [name]: {
                        multi_select: exampleOptions.map((n) => ({ name: n }))
                    }
                },
                description: multiOptions.length > 0
                    ? `Multiple selections from: ${multiOptionsList}`
                    : 'Multiple selections (no options defined yet)'
            };
        }
        case 'status': {
            const statusOptions = ((_e = propDef.status) === null || _e === void 0 ? void 0 : _e.options) || [];
            const firstStatus = ((_f = statusOptions[0]) === null || _f === void 0 ? void 0 : _f.name) || 'Status Name';
            const statusOptionsList = statusOptions.map((o) => o.name).join(', ');
            return {
                property_name: name,
                property_type: 'status',
                simple_value: firstStatus,
                notion_payload: {
                    [name]: { status: { name: firstStatus } }
                },
                description: statusOptions.length > 0
                    ? `Status from: ${statusOptionsList}`
                    : 'Status value (no options defined yet)'
            };
        }
        case 'date':
            return {
                property_name: name,
                property_type: 'date',
                simple_value: '2025-12-31',
                notion_payload: {
                    [name]: { date: { start: '2025-12-31' } }
                },
                description: 'ISO date (YYYY-MM-DD) or date range with end property'
            };
        case 'url':
            return {
                property_name: name,
                property_type: 'url',
                simple_value: 'https://example.com',
                notion_payload: {
                    [name]: { url: 'https://example.com' }
                },
                description: 'Valid URL starting with http:// or https://'
            };
        case 'email':
            return {
                property_name: name,
                property_type: 'email',
                simple_value: 'user@example.com',
                notion_payload: {
                    [name]: { email: 'user@example.com' }
                },
                description: 'Valid email address'
            };
        case 'phone_number':
            return {
                property_name: name,
                property_type: 'phone_number',
                simple_value: '+1-555-123-4567',
                notion_payload: {
                    [name]: { phone_number: '+1-555-123-4567' }
                },
                description: 'Phone number (any format)'
            };
        case 'people':
            return {
                property_name: name,
                property_type: 'people',
                simple_value: ['user-id-1', 'user-id-2'],
                notion_payload: {
                    [name]: {
                        people: [
                            { id: 'user-id-1' },
                            { id: 'user-id-2' }
                        ]
                    }
                },
                description: 'Array of Notion user IDs (use workspace users list to get IDs)'
            };
        case 'files':
            return {
                property_name: name,
                property_type: 'files',
                simple_value: 'https://example.com/file.pdf',
                notion_payload: {
                    [name]: {
                        files: [
                            {
                                name: 'file.pdf',
                                type: 'external',
                                external: { url: 'https://example.com/file.pdf' }
                            }
                        ]
                    }
                },
                description: 'External file URLs (Notion-hosted files cannot be set via API)'
            };
        case 'relation': {
            const relatedDbId = ((_g = propDef.relation) === null || _g === void 0 ? void 0 : _g.database_id) || 'related-database-id';
            return {
                property_name: name,
                property_type: 'relation',
                simple_value: ['page-id-1', 'page-id-2'],
                notion_payload: {
                    [name]: {
                        relation: [
                            { id: 'page-id-1' },
                            { id: 'page-id-2' }
                        ]
                    }
                },
                description: `Array of page IDs from related database (${relatedDbId})`
            };
        }
        // Read-only property types (cannot be set via API)
        case 'created_time':
            return {
                property_name: name,
                property_type: 'created_time',
                simple_value: null,
                notion_payload: {},
                description: 'Read-only: Automatically set when page is created'
            };
        case 'created_by':
            return {
                property_name: name,
                property_type: 'created_by',
                simple_value: null,
                notion_payload: {},
                description: 'Read-only: Automatically set to user who created the page'
            };
        case 'last_edited_time':
            return {
                property_name: name,
                property_type: 'last_edited_time',
                simple_value: null,
                notion_payload: {},
                description: 'Read-only: Automatically updated when page is edited'
            };
        case 'last_edited_by':
            return {
                property_name: name,
                property_type: 'last_edited_by',
                simple_value: null,
                notion_payload: {},
                description: 'Read-only: Automatically set to user who last edited the page'
            };
        case 'formula': {
            const expression = ((_h = propDef.formula) === null || _h === void 0 ? void 0 : _h.expression) || 'unknown';
            return {
                property_name: name,
                property_type: 'formula',
                simple_value: null,
                notion_payload: {},
                description: `Read-only: Computed formula (${expression})`
            };
        }
        case 'rollup': {
            const rollupFunc = ((_j = propDef.rollup) === null || _j === void 0 ? void 0 : _j.function) || 'unknown';
            return {
                property_name: name,
                property_type: 'rollup',
                simple_value: null,
                notion_payload: {},
                description: `Read-only: Rollup aggregation (${rollupFunc})`
            };
        }
        case 'unique_id':
            return {
                property_name: name,
                property_type: 'unique_id',
                simple_value: null,
                notion_payload: {},
                description: 'Read-only: Auto-incrementing unique ID'
            };
        case 'verification':
            return {
                property_name: name,
                property_type: 'verification',
                simple_value: null,
                notion_payload: {},
                description: 'Read-only: Verification status'
            };
        default:
            // Unsupported or unknown type
            return {
                property_name: name,
                property_type: type,
                simple_value: null,
                notion_payload: {},
                description: `Unsupported property type: ${type}`
            };
    }
}
/**
 * Format examples for human-readable console output
 *
 * @param examples - Array of property examples
 * @returns Formatted string
 */
function formatExamplesForConsole(examples) {
    const lines = [];
    lines.push('');
    lines.push('📋 Property Examples');
    lines.push('='.repeat(80));
    for (const example of examples) {
        lines.push('');
        lines.push(`${example.property_name} (${example.property_type})`);
        lines.push(`  ${example.description}`);
        if (example.simple_value !== null) {
            lines.push('');
            lines.push('  Simple value:');
            lines.push(`  ${JSON.stringify(example.simple_value)}`);
            lines.push('');
            lines.push('  Notion API payload:');
            const payload = JSON.stringify(example.notion_payload, null, 2);
            const indentedPayload = payload.split('\n').map(line => `  ${line}`).join('\n');
            lines.push(indentedPayload);
        }
        else {
            lines.push('');
            lines.push('  ⚠️  This property is read-only and cannot be set via API');
        }
        lines.push('-'.repeat(80));
    }
    return lines.join('\n');
}
/**
 * Group examples by writability (writable vs read-only)
 *
 * @param examples - Array of property examples
 * @returns Grouped examples
 */
function groupExamplesByWritability(examples) {
    const writable = [];
    const readOnly = [];
    for (const example of examples) {
        if (example.simple_value === null && Object.keys(example.notion_payload).length === 0) {
            readOnly.push(example);
        }
        else {
            writable.push(example);
        }
    }
    return { writable, readOnly };
}
