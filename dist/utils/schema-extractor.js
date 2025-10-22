"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.validateAgainstSchema = exports.formatSchemaAsMarkdown = exports.formatSchemaForTable = exports.filterProperties = exports.extractSchema = void 0;
/**
 * Extract clean, AI-parseable schema from Notion data source response
 *
 * This transforms the complex nested Notion API structure into a flat,
 * easy-to-understand format that AI agents can work with directly.
 *
 * @param dataSource - Raw Notion data source response
 * @returns Simplified schema object
 */
function extractSchema(dataSource) {
    const properties = [];
    // Extract title from data source
    const title = extractTitle(dataSource);
    // Extract description if available
    const description = extractDescription(dataSource);
    // Process each property in the data source
    if (dataSource.properties) {
        for (const [propName, propConfig] of Object.entries(dataSource.properties)) {
            const schema = extractPropertySchema(propName, propConfig);
            if (schema) {
                properties.push(schema);
            }
        }
    }
    return {
        id: dataSource.id,
        title,
        description,
        properties,
        url: 'url' in dataSource ? dataSource.url : undefined,
    };
}
exports.extractSchema = extractSchema;
/**
 * Extract title from data source
 */
function extractTitle(dataSource) {
    if ('title' in dataSource && Array.isArray(dataSource.title)) {
        return dataSource.title
            .map((t) => t.plain_text || '')
            .join('')
            .trim() || 'Untitled';
    }
    return 'Untitled';
}
/**
 * Extract description from data source
 */
function extractDescription(dataSource) {
    if ('description' in dataSource && Array.isArray(dataSource.description)) {
        const desc = dataSource.description
            .map((d) => d.plain_text || '')
            .join('')
            .trim();
        return desc || undefined;
    }
    return undefined;
}
/**
 * Extract individual property schema
 */
function extractPropertySchema(name, config) {
    var _a, _b, _c, _d, _e, _f, _g, _h, _j, _k;
    if (!config || !config.type) {
        return null;
    }
    const schema = {
        name,
        type: config.type,
    };
    // Handle select and multi-select with options
    if (config.type === 'select' && ((_a = config.select) === null || _a === void 0 ? void 0 : _a.options)) {
        schema.options = config.select.options.map((opt) => opt.name);
        schema.description = `Select one: ${schema.options.join(', ')}`;
    }
    if (config.type === 'multi_select' && ((_b = config.multi_select) === null || _b === void 0 ? void 0 : _b.options)) {
        schema.options = config.multi_select.options.map((opt) => opt.name);
        schema.description = `Select multiple: ${schema.options.join(', ')}`;
    }
    // Handle status property (similar to select)
    if (config.type === 'status' && ((_c = config.status) === null || _c === void 0 ? void 0 : _c.options)) {
        schema.options = config.status.options.map((opt) => opt.name);
        schema.description = `Status: ${schema.options.join(', ')}`;
    }
    // Handle formula properties
    if (config.type === 'formula' && ((_d = config.formula) === null || _d === void 0 ? void 0 : _d.expression)) {
        schema.config = {
            expression: config.formula.expression,
        };
        schema.description = `Formula: ${config.formula.expression}`;
    }
    // Handle rollup properties
    if (config.type === 'rollup') {
        schema.config = {
            relation_property: (_e = config.rollup) === null || _e === void 0 ? void 0 : _e.relation_property_name,
            rollup_property: (_f = config.rollup) === null || _f === void 0 ? void 0 : _f.rollup_property_name,
            function: (_g = config.rollup) === null || _g === void 0 ? void 0 : _g.function,
        };
        schema.description = 'Rollup from related database';
    }
    // Handle relation properties
    if (config.type === 'relation') {
        schema.config = {
            database_id: (_h = config.relation) === null || _h === void 0 ? void 0 : _h.database_id,
            type: (_j = config.relation) === null || _j === void 0 ? void 0 : _j.type,
        };
        schema.description = 'Relation to another database';
    }
    // Handle number properties with format
    if (config.type === 'number' && ((_k = config.number) === null || _k === void 0 ? void 0 : _k.format)) {
        schema.config = {
            format: config.number.format,
        };
        schema.description = `Number (${config.number.format})`;
    }
    // Mark title property as required
    if (config.type === 'title') {
        schema.required = true;
        schema.description = 'Title (required)';
    }
    return schema;
}
/**
 * Filter properties by names
 *
 * @param schema - Full schema
 * @param propertyNames - Array of property names to include
 * @returns Filtered schema
 */
function filterProperties(schema, propertyNames) {
    const lowerNames = propertyNames.map(n => n.toLowerCase());
    return {
        ...schema,
        properties: schema.properties.filter(p => lowerNames.includes(p.name.toLowerCase())),
    };
}
exports.filterProperties = filterProperties;
/**
 * Format schema as human-readable table data
 *
 * @param schema - Schema to format
 * @returns Array of objects for table display
 */
function formatSchemaForTable(schema) {
    return schema.properties.map(prop => {
        var _a;
        return ({
            name: prop.name,
            type: prop.type,
            required: prop.required ? 'Yes' : 'No',
            options: ((_a = prop.options) === null || _a === void 0 ? void 0 : _a.join(', ')) || '-',
            description: prop.description || '-',
        });
    });
}
exports.formatSchemaForTable = formatSchemaForTable;
/**
 * Format schema as markdown documentation
 *
 * @param schema - Schema to format
 * @returns Markdown string
 */
function formatSchemaAsMarkdown(schema) {
    var _a;
    const lines = [];
    lines.push(`# ${schema.title}`);
    lines.push('');
    if (schema.description) {
        lines.push(schema.description);
        lines.push('');
    }
    lines.push(`**Database ID:** \`${schema.id}\``);
    if (schema.url) {
        lines.push(`**URL:** ${schema.url}`);
    }
    lines.push('');
    lines.push('## Properties');
    lines.push('');
    lines.push('| Name | Type | Required | Options/Details |');
    lines.push('|------|------|----------|-----------------|');
    for (const prop of schema.properties) {
        const required = prop.required ? '✓' : '';
        const details = ((_a = prop.options) === null || _a === void 0 ? void 0 : _a.join(', ')) || prop.description || '';
        lines.push(`| ${prop.name} | ${prop.type} | ${required} | ${details} |`);
    }
    return lines.join('\n');
}
exports.formatSchemaAsMarkdown = formatSchemaAsMarkdown;
/**
 * Validate that a data object matches the schema
 *
 * @param schema - Schema to validate against
 * @param data - Data object to validate
 * @returns Validation result with errors
 */
function validateAgainstSchema(schema, data) {
    const errors = [];
    // Check required properties
    for (const prop of schema.properties) {
        if (prop.required && !(prop.name in data)) {
            errors.push(`Missing required property: ${prop.name}`);
        }
    }
    // Check property types and options
    for (const [key, value] of Object.entries(data)) {
        const propSchema = schema.properties.find(p => p.name === key);
        if (!propSchema) {
            errors.push(`Unknown property: ${key}`);
            continue;
        }
        // Validate select/multi-select options
        if (propSchema.options && propSchema.options.length > 0) {
            if (propSchema.type === 'select') {
                if (typeof value === 'string' && !propSchema.options.includes(value)) {
                    errors.push(`Invalid option for ${key}: ${value}. Must be one of: ${propSchema.options.join(', ')}`);
                }
            }
            if (propSchema.type === 'multi_select') {
                if (Array.isArray(value)) {
                    const invalidOptions = value.filter(v => !propSchema.options.includes(v));
                    if (invalidOptions.length > 0) {
                        errors.push(`Invalid options for ${key}: ${invalidOptions.join(', ')}`);
                    }
                }
            }
        }
    }
    return {
        valid: errors.length === 0,
        errors,
    };
}
exports.validateAgainstSchema = validateAgainstSchema;
