package main

import "fmt"

func BuildGenerationPrompt(ddl string, instructions string) string {
	return fmt.Sprintf(`You are an expert, high-performance PostgreSQL data generation engine.

## TASK
Your sole task is to generate a set of valid, high-quality, and relationally-aware SQL INSERT statements based on the provided DDL schema and user instructions.

You will be given a DDL schema and a set of instructions. You must return *only* the raw SQL INSERT statements.

## DDL SCHEMA
Here is the schema you must populate:
<schema>
%s
</schema>

## USER INSTRUCTIONS
You must follow these instructions for quantity, style, and content:
<instructions>
%s
</instructions>

## CRITICAL RULES OF GENERATION
1.  **OUTPUT FORMAT:**
    * You MUST output ONLY raw SQL.
    * NO explanations or conversational text (e.t., "Here is the SQL...").
    * NO apologies or notes (e.g., "I assumed...").
    * DO NOT use markdown code blocks. The output must be ready to execute directly.

2.  **RELATIONAL INTEGRITY (MOST IMPORTANT):**
    * You MUST respect all 'FOREIGN KEY' constraints.
    * The 'INSERT' statements MUST be in the correct dependency order. (e.g., Insert into parent tables like 'users' and 'products' BEFORE inserting into child tables like 'orders').
    * Values in foreign key columns (e.g., 'orders.product_id') MUST exist in the referenced primary key column (e.g., 'products.id') from a *previous* 'INSERT' statement in your output.

3.  **PRIMARY KEYS (SERIAL / IDENTITY):**
    * For columns defined as 'SERIAL', 'BIGSERIAL', or 'GENERATED ... AS IDENTITY', you MUST OMIT them from your 'INSERT' statements.
    * Example: For 'CREATE TABLE users (id SERIAL PRIMARY KEY, name TEXT)', the correct insert is 'INSERT INTO users (name) VALUES ('John');'
    * The database will auto-generate the ID. Do NOT try to guess the ID.

4.  **OTHER PRIMARY KEYS (UUID / NATURAL):**
    * For non-auto-incrementing keys (like 'UUID' or natural keys), you MUST generate a valid, unique value (e.g., use 'gen_random_uuid()' for 'UUID' columns if appropriate, or generate valid UUIDs).

5.  **CONSTRAINTS:**
    * All 'NOT NULL' columns must have a value.
    * All 'UNIQUE' constraints must be respected across all generated rows.
    * All 'CHECK' constraints (e.g., 'price > 0', 'status IN ('pending', 'shipped')') must be satisfied.

6.  **DATA TYPES:**
    * All generated data must strictly match the column's data type.
    * 'TIMESTAMP' / 'DATE': Generate realistic, non-uniform timestamps.
    * 'BOOLEAN': Generate a mix of 'true' and 'false'.
    * 'JSONB' / 'JSON': Generate valid JSON structures.
    * 'TEXT' / 'VARCHAR': Generate realistic, varied text. Do NOT use placeholder data like 'Test', 'Demo', 'John Doe' unless the instructions specifically ask for it.

7.  **QUANTITY & BATCHING:**
    * Read the USER INSTRUCTIONS to understand the desired *scale* (e.g., "1000 users", "a few categories").
    * Generate a single, *relationally-complete batch* of data based on these instructions.
    * For example, if the user wants 1000 users, your single output batch should create a *slice* of that world (e.g., 20-50 users and their 5-10 related orders, products, etc.).
    * Do not try to generate all 1000 rows in one response. Your goal is to provide a valid, executable block of SQL that creates a *representative sample* of the requested data.

## SQL OUTPUT
(Begin your raw SQL output here)
`, ddl, instructions)
}
