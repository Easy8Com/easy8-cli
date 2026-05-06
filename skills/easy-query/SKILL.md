---
name: easy-query
description: "Use when creating, extending, reviewing, or debugging any EasyQuery class. Covers filters, columns, sorting, grouping, registration, and dashboard availability."
---

# EasyQuery Skill (Easy8)

Use this skill when the user asks to:

- Create a new EasyQuery
- Add or remove filters or columns to an existing EasyQuery
- Debug a broken EasyQuery (missing records, wrong SQL, broken filters, broken sorting/grouping)
- Register a query for dashboards or EasyPage
- Understand how an existing EasyQuery works

## Primary Source of Truth (mandatory)

**Always read first:**

```
docs/backend_tutorials/features/easy_queries.md
```

This is the single authoritative source for EasyQuery design. Do not treat `app/easy_queries/easy_query.rb`
as design documentation — it is implementation and is not maintained as docs.

When examples are needed, inspect the nearest existing query implementation (same plugin or same model domain).
Prefer small, simple queries like `EasyWebHookQuery` or `EasyButtonQuery` as reference.

## Placement Rules

Decide where the new query belongs before writing any code:

| Situation                                                             | Location                           |
| --------------------------------------------------------------------- | ---------------------------------- |
| Query belongs to a feature that lives in a specific plugin/engine/rys | `<plugin>/app/easy_queries/`       |
| Query is truly cross-cutting platform behavior                        | `app/easy_queries/` (core)         |
| Client-specific query                                                 | `modification_*/app/easy_queries/` |

> Do not add a query to core if it belongs in a plugin — a customer may not have all plugins installed.

### Registration file

Registration uses the same pattern everywhere (core, plugin, engine, rys):

```ruby
EasyInitHelper.on_constant_autoload("EasyQuery") do
  EasyQuery.map do |query|
    query.register "MyModelEasyQuery"
  end
end
```

The registration file lives in `config/initializers/` of the owning component (plugin, engine, or core).

**File selection rule:**

1. Search the component's `config/initializers/` for an existing file that already contains `EasyInitHelper.on_constant_autoload("EasyQuery")`.
2. If found — add the `query.register` call inside the existing block.
3. If not found — create a new dedicated file, e.g. `config/initializers/easy_query_registration.rb`.

Do not invent a new file if a suitable one already exists. Do not use `ActiveSupport.on_load` or `Rails.application.config.after_initialize` for query registration — those are older patterns.

## Authoring Workflow

Follow this checklist when creating a new EasyQuery:

1. **Read** `docs/backend_tutorials/features/easy_queries.md` in full.
2. **Decide placement** (see Placement Rules above).
3. **Determine the class name**: `<ModelName>EasyQuery` (e.g. `MyModelEasyQuery`).
   Do not namespace unless strictly required — namespaced queries are not supported by `EasyGraphql`.
4. **Create the query file** in the correct `app/easy_queries/` directory.

Minimum required implementation:

```ruby
class MyModelEasyQuery < EasyQuery
  self.queried_class = ::MyModel

  def initialize_available_filters
    # one entry per filterable attribute
    add_available_filter "name"
    add_available_filter "status",
                         type: :list,
                         values: proc { queried_class.statuses.map { |k, v| [queried_class.human_enum_name("status", k), v] } }
    add_principal_autocomplete_filter "author_id"
  end

  def initialize_available_columns
    tbl = queried_class.table_name

    add_column :id,     :integer, sortable: "#{tbl}.id"
    add_column :name,   :string,  sortable: "#{tbl}.name"
    add_column :status, :record,  sortable: "#{tbl}.status"
    add_principal_column :author
  end

  def default_list_columns
    super.presence || %w[name status]
  end
end
```

5. **Add filter types correctly** — see Filter Reference below.
6. **Add column options correctly** — see Column Reference below.
7. **Register the query** — find or create the correct initializer (see Placement Rules above):

```ruby
EasyInitHelper.on_constant_autoload("EasyQuery") do
  EasyQuery.map do |query|
    query.register "MyModelEasyQuery"
  end
end
```

8. **Verify** by checking Filter Reference and Column Reference against every field you added.

## Filter Reference

Source of truth: `docs/backend_tutorials/features/easy_queries.md` — section `initialize_available_filters`.

Quick lookup:

| Filter type          | Required options                               | Notes                                                                                   |
| -------------------- | ---------------------------------------------- | --------------------------------------------------------------------------------------- |
| `:string` (default)  | —                                              | Simple text match                                                                       |
| `:integer`           | —                                              | —                                                                                       |
| `:float`             | —                                              | —                                                                                       |
| `:boolean`           | —                                              | —                                                                                       |
| `:date_period`       | —                                              | Use this instead of `:date` or `:date_past` (both are deprecated)                       |
| `:list`              | `values:` (proc or array)                      | Static list                                                                             |
| `:list_optional`     | `values:`                                      | Like `:list` but value can be empty                                                     |
| `:list_autocomplete` | `source:`, `source_root:`, optionally `klass:` | Dynamic autocomplete; use `add_principal_autocomplete_filter` for Principal descendants |
| `:tree`              | `source:`, `source_root:`                      | Hierarchical autocomplete                                                               |
| `:text`              | —                                              | Multiline text                                                                          |

For `Principal` filters (`author_id`, `user_id`, `assignee_id`, etc.) always use the helper:

```ruby
add_principal_autocomplete_filter "author_id"
```

For `:list_autocomplete` with a non-Principal association:

```ruby
add_available_filter "collection_id",
                     type: :list_autocomplete,
                     source: "all_collections",
                     source_root: "entities",
                     klass: ::Collection
```

## Column Reference

Source of truth: `docs/backend_tutorials/features/easy_queries.md` — section `initialize_available_columns`.

Quick lookup of column types:

| Column type | Use for                |
| ----------- | ---------------------- |
| `:string`   | Text columns           |
| `:integer`  | Integer columns        |
| `:float`    | Float columns          |
| `:boolean`  | Boolean columns        |
| `:text`     | Multiline text         |
| `:date`     | Date columns           |
| `:record`   | Associations and enums |

Common column options:

| Option           | Required for                         | Value format                                                                   |
| ---------------- | ------------------------------------ | ------------------------------------------------------------------------------ |
| `sortable:`      | Sortable columns                     | SQL string e.g. `"#{tbl}.name"` or array for multi-level                       |
| `default_order:` | —                                    | `"asc"` or `"desc"`                                                            |
| `groupable:`     | Groupable columns                    | SQL string e.g. `"#{tbl}.status"` or `true` when column name matches DB column |
| `preload:`       | Association columns                  | Array of symbols e.g. `[:collection]`                                          |
| `includes:`      | Sortable-through-association columns | Array of symbols; required when `sortable` spans another table                 |
| `most_used:`     | —                                    | Boolean, default `false`                                                       |

For `Principal` columns always use the helper:

```ruby
add_principal_column :author
```

For association columns that need inline editing in the table:

```ruby
add_column :project, :record,
           sortable: "#{Project.table_name}.name",
           groupable: "#{Issue.table_name}.project_id",
           includes: [:project],
           attribute: "project_id", ref: :itself,
           source_options: { source: "allowed_target_projects_on_move", source_root: "projects" }
```

## Debugging Workflow

When an EasyQuery is broken, follow these steps in order before reading internals:

### Step 1 — Identify the symptom

| Symptom                                      | Most likely cause                                                                                               |
| -------------------------------------------- | --------------------------------------------------------------------------------------------------------------- |
| Query not available in dashboards / EasyPage | Not registered, or registration uses wrong wrapper (must be `EasyInitHelper.on_constant_autoload("EasyQuery")`) |
| Filter not visible in UI                     | Filter not added in `initialize_available_filters`                                                              |
| Filter produces no results or wrong SQL      | Wrong filter type, missing `values` / `source` / `source_root`                                                  |
| Column missing from column picker            | Column not added in `initialize_available_columns`                                                              |
| Sorting broken                               | Missing or wrong `sortable:` option, or missing `includes:` for cross-table sort                                |
| Grouping broken                              | Missing or wrong `groupable:` option                                                                            |
| N+1 or missing preloads                      | Missing `preload:` or `includes:` on column definition                                                          |
| Association column shows nil                 | Missing `preload:` on column                                                                                    |
| Wrong records returned                       | Wrong `self.queried_class`, wrong `statement`, wrong filter field name vs DB column                             |

### Step 2 — Check docs first

Re-read `docs/backend_tutorials/features/easy_queries.md` for the relevant section.
Cross-check each filter and column definition against Filter Reference and Column Reference tables above.

### Step 3 — Compare to a nearby working query

Find the simplest existing query for a similar model in the same plugin or domain.
Compare filter/column definitions field by field.

### Step 4 — Use the debug cookbook

See Debug Cookbook below.

### Step 5 — Only if the above is not enough

Inspect the query class file directly. Do not read `app/easy_queries/easy_query.rb` as primary reference.

## Debug Cookbook

Minimal Rails console commands for verifying a query:

```ruby
# Instantiate a fresh query
q = MyModelEasyQuery.new

# Check available filter keys
q.available_filters.keys
# => ["name", "status", "author_id", ...]

# Check available column names
q.available_columns.map(&:name)
# => [:id, :name, :status, :author, ...]

# Check default list columns
q.default_list_columns
# => ["name", "status"]

# Check the generated SQL WHERE clause (with no filters set)
q.statement
# => nil or a SQL fragment

# Apply a filter and inspect the statement
q.filters = { "status" => { operator: "=", values: ["1"] } }
q.statement
# => "my_models.status = 1"

# Inspect the full scope SQL
q.new_entity_scope.to_sql

# Check sorting SQL
q.sort_criteria = [["name", "asc"]]
q.sort_criteria_to_sql_order
# => "my_models.name asc, my_models.id ASC"

# Check registration
EasyQuery.registered_subclasses.keys
# => [..., "MyModelEasyQuery", ...]
```

## Common Mistakes

- Query not registered, or registration uses old `ActiveSupport.on_load` / `after_initialize` pattern instead of `EasyInitHelper.on_constant_autoload("EasyQuery")`
- Registration added to a wrong component's initializer (e.g. core initializer for a plugin query)
- New registration file created when an existing initializer in the same component already has the `EasyInitHelper.on_constant_autoload("EasyQuery")` block

## Self-Check Before Finalizing

Before marking EasyQuery work complete, verify:

- [ ] `docs/backend_tutorials/features/easy_queries.md` was read first
- [ ] Query placed in the correct plugin/engine/core location
- [ ] `self.queried_class` is set
- [ ] `initialize_available_filters` implemented with correct types and required options per filter
- [ ] `initialize_available_columns` implemented with correct types and options
- [ ] `default_list_columns` returns array of strings (not symbols)
- [ ] Query registered using `EasyInitHelper.on_constant_autoload("EasyQuery")` in the correct initializer
- [ ] Existing initializer reused if it already contains the `on_constant_autoload("EasyQuery")` block
- [ ] Association columns have `preload:` or `includes:` where needed
- [ ] `sortable:` columns that cross tables have corresponding `includes:`
- [ ] No namespacing if EasyGraphql usage is expected
- [ ] Debug cookbook commands run and produce expected output
