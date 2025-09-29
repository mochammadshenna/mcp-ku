-- Drop triggers first
DROP TRIGGER IF EXISTS update_tools_updated_at ON tools;
DROP TRIGGER IF EXISTS update_evaluations_updated_at ON evaluations;
DROP TRIGGER IF EXISTS update_prompts_updated_at ON prompts;
DROP TRIGGER IF EXISTS update_flows_updated_at ON flows;
DROP TRIGGER IF EXISTS update_mcp_servers_updated_at ON mcp_servers;

-- Drop function
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop tables in reverse order (respecting foreign key constraints)
DROP TABLE IF EXISTS tool_executions;
DROP TABLE IF EXISTS tools;
DROP TABLE IF EXISTS evaluation_results;
DROP TABLE IF EXISTS evaluations;
DROP TABLE IF EXISTS generations;
DROP TABLE IF EXISTS vector_documents;
DROP TABLE IF EXISTS prompts;
DROP TABLE IF EXISTS flow_executions;
DROP TABLE IF EXISTS flows;
DROP TABLE IF EXISTS mcp_servers;

-- Drop extension
DROP EXTENSION IF EXISTS vector;