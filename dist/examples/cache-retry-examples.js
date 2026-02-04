"use strict";
/**
 * Examples demonstrating enhanced retry logic and caching layer
 *
 * This file provides practical examples of how to use the new features.
 * These examples can be adapted to your specific use cases.
 */
Object.defineProperty(exports, "__esModule", { value: true });
exports.example1_basicUsage = example1_basicUsage;
exports.example2_cacheStats = example2_cacheStats;
exports.example3_cacheInvalidation = example3_cacheInvalidation;
exports.example4_customRetry = example4_customRetry;
exports.example5_circuitBreaker = example5_circuitBreaker;
exports.example6_batchOperations = example6_batchOperations;
exports.example7_errorCategorization = example7_errorCategorization;
exports.example8_delayCalculation = example8_delayCalculation;
exports.example9_productionPattern = example9_productionPattern;
exports.example10_configurationShowcase = example10_configurationShowcase;
exports.runAllExamples = runAllExamples;
const notion = require("../notion");
const cache_1 = require("../cache");
const retry_1 = require("../retry");
/**
 * Example 1: Basic usage with automatic caching and retry
 */
async function example1_basicUsage() {
    console.log('\n=== Example 1: Basic Usage ===');
    const databaseId = 'your-database-id';
    try {
        // First call - will cache the result
        console.log('First call (cache MISS expected):');
        const ds1 = await notion.retrieveDataSource(databaseId);
        console.log(`Retrieved data source: ${ds1.id}`);
        // Second call - will use cache
        console.log('\nSecond call (cache HIT expected):');
        const ds2 = await notion.retrieveDataSource(databaseId);
        console.log(`Retrieved data source: ${ds2.id}`);
        // Verify both are the same (cached)
        console.log(`\nSame object from cache: ${ds1 === ds2}`);
    }
    catch (error) {
        console.error('Error:', error.message);
    }
}
/**
 * Example 2: Monitoring cache performance
 */
async function example2_cacheStats() {
    console.log('\n=== Example 2: Cache Statistics ===');
    // Clear cache to start fresh
    cache_1.cacheManager.clear();
    console.log('Cache cleared');
    const databaseIds = ['db-id-1', 'db-id-2', 'db-id-3'];
    // Make multiple calls
    for (let i = 0; i < 3; i++) {
        console.log(`\nRound ${i + 1}:`);
        for (const dbId of databaseIds) {
            try {
                await notion.retrieveDataSource(dbId);
                console.log(`  Retrieved ${dbId}`);
            }
            catch {
                console.log(`  Failed to retrieve ${dbId}`);
            }
        }
    }
    // Display cache statistics
    const stats = cache_1.cacheManager.getStats();
    const hitRate = cache_1.cacheManager.getHitRate();
    console.log('\n=== Cache Statistics ===');
    console.log(`Total Requests: ${stats.hits + stats.misses}`);
    console.log(`Cache Hits: ${stats.hits}`);
    console.log(`Cache Misses: ${stats.misses}`);
    console.log(`Hit Rate: ${(hitRate * 100).toFixed(2)}%`);
    console.log(`Current Size: ${stats.size} entries`);
    console.log(`Evictions: ${stats.evictions}`);
}
/**
 * Example 3: Manual cache invalidation
 */
async function example3_cacheInvalidation() {
    console.log('\n=== Example 3: Cache Invalidation ===');
    const databaseId = 'your-database-id';
    try {
        // Retrieve and cache
        console.log('Initial retrieval:');
        const ds1 = await notion.retrieveDataSource(databaseId);
        console.log(`Retrieved: ${ds1.id}`);
        // Update the database
        console.log('\nUpdating database...');
        await notion.updateDataSource({
            data_source_id: databaseId,
            title: [{ type: 'text', text: { content: 'Updated Title' } }]
        });
        console.log('Database updated (cache automatically invalidated)');
        // Retrieve again - will fetch fresh data
        console.log('\nRetrieving after update:');
        const ds2 = await notion.retrieveDataSource(databaseId);
        console.log(`Retrieved: ${ds2.id}`);
        // Manual invalidation example
        console.log('\nManual cache invalidation:');
        cache_1.cacheManager.invalidate('dataSource', databaseId);
        console.log('Cache invalidated for specific data source');
        // Or invalidate all data sources
        cache_1.cacheManager.invalidate('dataSource');
        console.log('Cache invalidated for all data sources');
    }
    catch (error) {
        console.error('Error:', error.message);
    }
}
/**
 * Example 4: Custom retry configuration
 */
async function example4_customRetry() {
    console.log('\n=== Example 4: Custom Retry Configuration ===');
    try {
        const result = await (0, retry_1.fetchWithRetry)(async () => {
            console.log('Attempting API call...');
            // Simulate an operation that might fail
            return await notion.client.users.me({});
        }, {
            config: {
                maxRetries: 5,
                baseDelay: 2000, // Start with 2 second delay
                maxDelay: 60000, // Cap at 60 seconds
                exponentialBase: 2.5, // Increase delay by 2.5x each time
                jitterFactor: 0.2, // Add 20% random variation
            },
            onRetry: (context) => {
                console.log(`Retry ${context.attempt}/${context.maxRetries}: ` +
                    `Last error: ${context.lastError.code || context.lastError.status}. ` +
                    `Total delay so far: ${context.totalDelay}ms`);
            },
            context: 'getCriticalUserInfo'
        });
        console.log('Success:', result);
    }
    catch (error) {
        console.error('Final error after all retries:', error.message);
    }
}
/**
 * Example 5: Circuit breaker pattern
 */
async function example5_circuitBreaker() {
    console.log('\n=== Example 5: Circuit Breaker Pattern ===');
    const breaker = new retry_1.CircuitBreaker(3, // Open circuit after 3 failures
    2, // Close after 2 successes
    30000 // 30 second timeout
    );
    const databaseIds = ['bad-id-1', 'bad-id-2', 'bad-id-3', 'bad-id-4', 'bad-id-5'];
    for (const dbId of databaseIds) {
        try {
            console.log(`\nAttempting to retrieve ${dbId}...`);
            const state = breaker.getState();
            console.log(`Circuit breaker state: ${state.state} (failures: ${state.failures})`);
            const result = await breaker.execute(() => notion.retrieveDataSource(dbId));
            console.log(`Success: ${result.id}`);
        }
        catch (error) {
            console.error(`Failed: ${error.message}`);
            const state = breaker.getState();
            if (state.state === 'open') {
                console.error('Circuit breaker is OPEN - stopping further attempts');
                break;
            }
        }
    }
    // Show final state
    const finalState = breaker.getState();
    console.log('\nFinal circuit breaker state:', finalState);
}
/**
 * Example 6: Batch operations with retry
 */
async function example6_batchOperations() {
    console.log('\n=== Example 6: Batch Operations with Retry ===');
    const databaseIds = ['db-1', 'db-2', 'db-3', 'db-4', 'db-5'];
    const operations = databaseIds.map(dbId => () => notion.retrieveDataSource(dbId));
    console.log('Processing batch with concurrency limit...');
    const results = await (0, retry_1.batchWithRetry)(operations, {
        concurrency: 2, // Process 2 at a time
        config: {
            maxRetries: 3,
            baseDelay: 1000,
        },
        onRetry: (context) => {
            console.log(`  Retry ${context.attempt} for operation`);
        }
    });
    // Analyze results
    const successful = results.filter(r => r.success).length;
    const failed = results.filter(r => !r.success).length;
    console.log('\n=== Batch Results ===');
    console.log(`Total operations: ${results.length}`);
    console.log(`Successful: ${successful}`);
    console.log(`Failed: ${failed}`);
    // Show details for failed operations
    if (failed > 0) {
        console.log('\nFailed operations:');
        results.forEach((result, index) => {
            if (!result.success) {
                console.log(`  Operation ${index + 1}: ${result.error.message}`);
            }
        });
    }
}
/**
 * Example 7: Error categorization
 */
async function example7_errorCategorization() {
    console.log('\n=== Example 7: Error Categorization ===');
    // Simulate different types of errors
    const errors = [
        { status: 429, message: 'Rate limited' },
        { status: 500, message: 'Internal server error' },
        { status: 400, message: 'Bad request' },
        { status: 401, message: 'Unauthorized' },
        { status: 503, message: 'Service unavailable' },
        { code: 'ECONNRESET', message: 'Connection reset' },
        { code: 'ETIMEDOUT', message: 'Timeout' },
    ];
    console.log('Checking which errors are retryable:\n');
    for (const error of errors) {
        const retryable = (0, retry_1.isRetryableError)(error);
        const status = error.status ? `HTTP ${error.status}` : error.code;
        console.log(`${status} - ${error.message}: ` +
            `${retryable ? 'RETRYABLE ✓' : 'NON-RETRYABLE ✗'}`);
    }
}
/**
 * Example 8: Delay calculation visualization
 */
async function example8_delayCalculation() {
    console.log('\n=== Example 8: Delay Calculation ===');
    const configs = [
        { name: 'Default', baseDelay: 1000, exponentialBase: 2, jitterFactor: 0.1 },
        { name: 'Aggressive', baseDelay: 2000, exponentialBase: 3, jitterFactor: 0.2 },
        { name: 'Conservative', baseDelay: 500, exponentialBase: 1.5, jitterFactor: 0.05 },
    ];
    for (const config of configs) {
        console.log(`\n${config.name} configuration:`);
        console.log(`Base: ${config.baseDelay}ms, Exponential: ${config.exponentialBase}, Jitter: ${config.jitterFactor}`);
        console.log('Retry delays:');
        for (let attempt = 1; attempt <= 5; attempt++) {
            const delay = (0, retry_1.calculateDelay)(attempt, {
                maxRetries: 5,
                baseDelay: config.baseDelay,
                maxDelay: 30000,
                exponentialBase: config.exponentialBase,
                jitterFactor: config.jitterFactor,
                retryableStatusCodes: [],
                retryableErrorCodes: []
            });
            console.log(`  Attempt ${attempt}: ~${delay}ms`);
        }
    }
}
/**
 * Example 9: Production-ready pattern
 */
async function example9_productionPattern() {
    var _a, _b;
    console.log('\n=== Example 9: Production-Ready Pattern ===');
    // Setup circuit breaker for resilience
    const breaker = new retry_1.CircuitBreaker(10, 3, 120000);
    // Helper function with comprehensive error handling
    async function safeDataSourceRetrieval(dataSourceId) {
        try {
            const result = await breaker.execute(() => notion.retrieveDataSource(dataSourceId), {
                config: {
                    maxRetries: 5,
                    baseDelay: 1000,
                    maxDelay: 30000,
                },
                onRetry: (context) => {
                    // Log to monitoring service
                    console.log(`[RETRY] DataSource ${dataSourceId}: ` +
                        `attempt ${context.attempt}/${context.maxRetries}`);
                },
                context: `retrieveDataSource:${dataSourceId}`
            });
            return { success: true, data: result, error: null };
        }
        catch (error) {
            // Log to error tracking service
            console.error(`[ERROR] Failed to retrieve data source ${dataSourceId}: ${error.message}`);
            // Check circuit breaker state
            const state = breaker.getState();
            if (state.state === 'open') {
                console.error('[CRITICAL] Circuit breaker is open - service may be down');
            }
            return { success: false, data: null, error };
        }
    }
    // Usage
    const databaseId = 'your-database-id';
    console.log(`Retrieving data source: ${databaseId}`);
    const result = await safeDataSourceRetrieval(databaseId);
    if (result.success) {
        console.log('Success:', (_a = result.data) === null || _a === void 0 ? void 0 : _a.id);
        // Get cache statistics for monitoring
        // Cache statistics available via cacheManager.getStats() if needed
        console.log(`Cache hit rate: ${(cache_1.cacheManager.getHitRate() * 100).toFixed(2)}%`);
    }
    else {
        console.error('Operation failed:', (_b = result.error) === null || _b === void 0 ? void 0 : _b.message);
    }
}
/**
 * Example 10: Configuration showcase
 */
async function example10_configurationShowcase() {
    console.log('\n=== Example 10: Configuration Showcase ===');
    // Show current cache configuration
    const cacheConfig = cache_1.cacheManager.getConfig();
    console.log('\nCache Configuration:');
    console.log(`  Enabled: ${cacheConfig.enabled}`);
    console.log(`  Default TTL: ${cacheConfig.defaultTtl}ms`);
    console.log(`  Max Size: ${cacheConfig.maxSize}`);
    console.log('  TTL by type:');
    console.log(`    Data Sources: ${cacheConfig.ttlByType.dataSource}ms`);
    console.log(`    Databases: ${cacheConfig.ttlByType.database}ms`);
    console.log(`    Users: ${cacheConfig.ttlByType.user}ms`);
    console.log(`    Pages: ${cacheConfig.ttlByType.page}ms`);
    console.log(`    Blocks: ${cacheConfig.ttlByType.block}ms`);
    // Show environment variables
    console.log('\nEnvironment Variables:');
    console.log(`  NOTION_CLI_MAX_RETRIES: ${process.env.NOTION_CLI_MAX_RETRIES || 'default (3)'}`);
    console.log(`  NOTION_CLI_BASE_DELAY: ${process.env.NOTION_CLI_BASE_DELAY || 'default (1000ms)'}`);
    console.log(`  NOTION_CLI_CACHE_ENABLED: ${process.env.NOTION_CLI_CACHE_ENABLED || 'default (true)'}`);
    console.log(`  DEBUG: ${process.env.DEBUG || 'false'}`);
}
/**
 * Run all examples
 */
async function runAllExamples() {
    console.log('='.repeat(60));
    console.log('Enhanced Retry Logic and Caching Examples');
    console.log('='.repeat(60));
    // Note: These examples assume you have a valid NOTION_TOKEN
    // and valid database IDs. Adjust the IDs in each example.
    await example1_basicUsage();
    await example2_cacheStats();
    await example3_cacheInvalidation();
    await example4_customRetry();
    await example5_circuitBreaker();
    await example6_batchOperations();
    await example7_errorCategorization();
    await example8_delayCalculation();
    await example9_productionPattern();
    await example10_configurationShowcase();
    console.log('\n' + '='.repeat(60));
    console.log('All examples completed!');
    console.log('='.repeat(60));
}
// Export all examples
exports.default = {
    example1_basicUsage,
    example2_cacheStats,
    example3_cacheInvalidation,
    example4_customRetry,
    example5_circuitBreaker,
    example6_batchOperations,
    example7_errorCategorization,
    example8_delayCalculation,
    example9_productionPattern,
    example10_configurationShowcase,
    runAllExamples,
};
// Run examples if executed directly
if (require.main === module) {
    runAllExamples().catch(console.error);
}
