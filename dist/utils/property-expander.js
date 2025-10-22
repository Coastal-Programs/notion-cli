"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.validateSimpleProperties = exports.expandSimpleProperties = void 0;
/**
 * Expand simple flat properties to Notion API format
 *
 * This function takes simplified property values and automatically expands them
 * to the correct Notion API structure based on the database schema.
 *
 * @param simple - Flat key-value property object
 * @param schema - Database properties schema from data source
 * @returns Properly formatted Notion properties object
 *
 * @example
 * // Input (simple):
 * { "Name": "My Task", "Status": "In Progress", "Priority": 5 }
 *
 * // Output (Notion format):
 * {
 *   "Name": { "title": [{ "text": { "content": "My Task" } }] },
 *   "Status": { "select": { "name": "In Progress" } },
 *   "Priority": { "number": 5 }
 * }
 */
async function expandSimpleProperties(simple, schema) {
    const expanded = {};
    for (const [propName, value] of Object.entries(simple)) {
        // Find property in schema (case-insensitive)
        const propDef = findProperty(schema, propName);
        if (!propDef) {
            throw new Error(`Property "${propName}" not found in database schema.\n` +
                `Available properties: ${Object.keys(schema).join(', ')}`);
        }
        // Expand based on type
        try {
            expanded[propDef.actualName] = expandProperty(value, propDef.type, propDef);
        }
        catch (error) {
            throw new Error(`Error expanding property "${propName}": ${error.message}`);
        }
    }
    return expanded;
}
exports.expandSimpleProperties = expandSimpleProperties;
/**
 * Find property in schema with case-insensitive matching
 */
function findProperty(schema, name) {
    const normalized = name.toLowerCase();
    for (const [key, value] of Object.entries(schema)) {
        if (key.toLowerCase() === normalized) {
            // Cast value to object type to allow spreading
            const propConfig = value;
            return { actualName: key, ...propConfig };
        }
    }
    return null;
}
/**
 * Expand a single property value to Notion format based on type
 */
function expandProperty(value, type, propDef) {
    // Handle null values
    if (value === null) {
        return null;
    }
    switch (type) {
        case 'title':
            return {
                title: [{ text: { content: String(value) } }]
            };
        case 'rich_text':
            return {
                rich_text: [{ text: { content: String(value) } }]
            };
        case 'number':
            const num = Number(value);
            if (isNaN(num)) {
                throw new Error(`Invalid number value: "${value}"`);
            }
            return { number: num };
        case 'checkbox':
            // Handle boolean or string representations
            let boolValue;
            if (typeof value === 'boolean') {
                boolValue = value;
            }
            else if (typeof value === 'string') {
                const lower = value.toLowerCase();
                if (lower === 'true' || lower === 'yes' || lower === '1') {
                    boolValue = true;
                }
                else if (lower === 'false' || lower === 'no' || lower === '0') {
                    boolValue = false;
                }
                else {
                    throw new Error(`Invalid checkbox value: "${value}". Use true/false, yes/no, or 1/0`);
                }
            }
            else {
                boolValue = Boolean(value);
            }
            return { checkbox: boolValue };
        case 'select':
            return expandSelectProperty(value, propDef);
        case 'multi_select':
            return expandMultiSelectProperty(value, propDef);
        case 'status':
            return expandStatusProperty(value, propDef);
        case 'date':
            return expandDateProperty(value);
        case 'url':
            const urlStr = String(value);
            // Basic URL validation
            if (!urlStr.match(/^https?:\/\/.+/)) {
                throw new Error(`Invalid URL: "${value}". Must start with http:// or https://`);
            }
            return { url: urlStr };
        case 'email':
            const emailStr = String(value);
            // Basic email validation
            if (!emailStr.match(/^[^\s@]+@[^\s@]+\.[^\s@]+$/)) {
                throw new Error(`Invalid email: "${value}"`);
            }
            return { email: emailStr };
        case 'phone_number':
            return { phone_number: String(value) };
        case 'people':
            return expandPeopleProperty(value);
        case 'files':
            // Files need external URLs
            const files = Array.isArray(value) ? value : [value];
            return {
                files: files.map(f => {
                    if (typeof f === 'string') {
                        return { name: f, external: { url: f } };
                    }
                    return f;
                })
            };
        case 'relation':
            // Relations need page IDs
            const relations = Array.isArray(value) ? value : [value];
            return {
                relation: relations.map(id => ({ id: String(id) }))
            };
        default:
            throw new Error(`Unsupported property type: ${type}. ` +
                `Supported types: title, rich_text, number, checkbox, select, multi_select, ` +
                `status, date, url, email, phone_number, people, files, relation`);
    }
}
/**
 * Expand select property with validation
 */
function expandSelectProperty(value, propDef) {
    var _a;
    const selectOptions = ((_a = propDef.select) === null || _a === void 0 ? void 0 : _a.options) || [];
    const strValue = String(value);
    // Case-insensitive matching
    const validOption = selectOptions.find((opt) => opt.name.toLowerCase() === strValue.toLowerCase());
    if (!validOption && selectOptions.length > 0) {
        const optionNames = selectOptions.map((o) => o.name).join(', ');
        throw new Error(`Invalid select value: "${value}"\n` +
            `Valid options: ${optionNames}\n` +
            `Tip: Values are case-insensitive`);
    }
    // Use the exact option name from schema (preserving case)
    const exactName = validOption ? validOption.name : strValue;
    return { select: { name: exactName } };
}
/**
 * Expand multi-select property with validation
 */
function expandMultiSelectProperty(value, propDef) {
    var _a;
    const values = Array.isArray(value) ? value : [value];
    const multiOptions = ((_a = propDef.multi_select) === null || _a === void 0 ? void 0 : _a.options) || [];
    const validated = values.map(v => {
        const strValue = String(v);
        // Case-insensitive matching
        const validOption = multiOptions.find((opt) => opt.name.toLowerCase() === strValue.toLowerCase());
        if (!validOption && multiOptions.length > 0) {
            const optionNames = multiOptions.map((o) => o.name).join(', ');
            throw new Error(`Invalid multi-select value: "${v}"\n` +
                `Valid options: ${optionNames}`);
        }
        // Use exact option name from schema
        const exactName = validOption ? validOption.name : strValue;
        return { name: exactName };
    });
    return { multi_select: validated };
}
/**
 * Expand status property with validation
 */
function expandStatusProperty(value, propDef) {
    var _a;
    const statusOptions = ((_a = propDef.status) === null || _a === void 0 ? void 0 : _a.options) || [];
    const strValue = String(value);
    // Case-insensitive matching
    const validStatus = statusOptions.find((opt) => opt.name.toLowerCase() === strValue.toLowerCase());
    if (!validStatus && statusOptions.length > 0) {
        const optionNames = statusOptions.map((o) => o.name).join(', ');
        throw new Error(`Invalid status value: "${value}"\n` +
            `Valid options: ${optionNames}`);
    }
    // Use exact status name from schema
    const exactName = validStatus ? validStatus.name : strValue;
    return { status: { name: exactName } };
}
/**
 * Expand date property with support for ISO dates and relative dates
 */
function expandDateProperty(value) {
    const dateStr = parseRelativeDate(String(value));
    // Check if it includes time (ISO 8601 with time component)
    if (dateStr.includes('T')) {
        return { date: { start: dateStr } };
    }
    return { date: { start: dateStr } };
}
/**
 * Parse relative date strings like "today", "tomorrow", "+7 days"
 */
function parseRelativeDate(value) {
    // Handle ISO dates (YYYY-MM-DD or full ISO 8601)
    if (/^\d{4}-\d{2}-\d{2}/.test(value)) {
        return value;
    }
    // Handle relative dates
    const today = new Date();
    today.setHours(0, 0, 0, 0); // Reset to start of day
    if (value.toLowerCase() === 'today') {
        return today.toISOString().split('T')[0];
    }
    if (value.toLowerCase() === 'tomorrow') {
        today.setDate(today.getDate() + 1);
        return today.toISOString().split('T')[0];
    }
    if (value.toLowerCase() === 'yesterday') {
        today.setDate(today.getDate() - 1);
        return today.toISOString().split('T')[0];
    }
    // Parse "+N days/weeks/months/years" format
    const match = value.match(/^([+-]?\d+)\s*(day|week|month|year)s?$/i);
    if (match) {
        const amount = parseInt(match[1]);
        const unit = match[2].toLowerCase();
        switch (unit) {
            case 'day':
                today.setDate(today.getDate() + amount);
                break;
            case 'week':
                today.setDate(today.getDate() + amount * 7);
                break;
            case 'month':
                today.setMonth(today.getMonth() + amount);
                break;
            case 'year':
                today.setFullYear(today.getFullYear() + amount);
                break;
        }
        return today.toISOString().split('T')[0];
    }
    // If none of the above, assume it's already a valid date string
    return value;
}
/**
 * Expand people property
 */
function expandPeopleProperty(value) {
    const users = Array.isArray(value) ? value : [value];
    return {
        people: users.map(u => {
            // Support user ID or email
            if (typeof u === 'string') {
                // Check if it's a UUID (user ID) or email
                if (u.match(/^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i)) {
                    return { id: u };
                }
                // For email, we can only use ID - throw helpful error
                if (u.includes('@')) {
                    throw new Error(`Cannot use email addresses for people property. ` +
                        `Use Notion user IDs instead. You can get user IDs with: notion-cli user list`);
                }
                return { id: u };
            }
            return { id: String(u) };
        })
    };
}
/**
 * Validate simple properties against schema before expansion
 * This can be called optionally before expandSimpleProperties to get detailed errors
 */
function validateSimpleProperties(simple, schema) {
    const errors = [];
    for (const [propName, value] of Object.entries(simple)) {
        const propDef = findProperty(schema, propName);
        if (!propDef) {
            errors.push(`Property "${propName}" not found in schema`);
            continue;
        }
        // Type-specific validation
        try {
            expandProperty(value, propDef.type, propDef);
        }
        catch (error) {
            errors.push(`${propName}: ${error.message}`);
        }
    }
    return {
        valid: errors.length === 0,
        errors
    };
}
exports.validateSimpleProperties = validateSimpleProperties;
