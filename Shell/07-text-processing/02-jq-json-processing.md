# jq and JSON Processing

## What Problem This Solves

Every API returns JSON. Kubernetes, Docker, cloud CLIs, GitHub — all JSON. You need to extract values, filter arrays, transform structures. grep and awk can parse simple JSON but break on nested objects, arrays, and multiline formatting. **jq** parses JSON correctly and is essential for modern shell work.

## Installing jq

```bash
sudo dnf install jq
```

## Mental Model

jq works like a pipeline — each filter takes input and produces output:

```
JSON Input → filter1 → filter2 → filter3 → Output
```

```bash
echo '{"name": "Alice", "age": 30}' | jq '.name'
# "Alice"

echo '{"name": "Alice", "age": 30}' | jq -r '.name'
# Alice (raw output — no quotes)
```

The `-r` flag gives **raw output** (strips quotes). You'll use it almost always.

## Basic Filters

### Identity and Field Access

```bash
# Identity — pretty-print:
curl -s https://api.example.com/data | jq '.'

# Field access:
echo '{"host": "db.internal", "port": 5432}' | jq '.host'
# "db.internal"

# Nested field:
echo '{"server": {"host": "db.internal", "port": 5432}}' | jq '.server.host'
# "db.internal"

# Optional field (no error if missing):
echo '{"a": 1}' | jq '.b'        # null
echo '{"a": 1}' | jq '.b // "default"'  # "default" (alternative operator)
```

### Array Access

```bash
data='["apple", "banana", "cherry"]'

echo "$data" | jq '.[0]'         # "apple"
echo "$data" | jq '.[-1]'        # "cherry" (last element)
echo "$data" | jq '.[1:3]'       # ["banana", "cherry"] (slice)
echo "$data" | jq '. | length'   # 3

# Array of objects:
pods='[{"name": "web", "status": "Running"}, {"name": "db", "status": "Pending"}]'
echo "$pods" | jq '.[0].name'    # "web"
echo "$pods" | jq '.[].name'     # "web" \n "db" (iterate all)
```

### Iterating Arrays with `.[]`

```bash
# .[] explodes an array into individual elements:
echo '[1, 2, 3]' | jq '.[]'
# 1
# 2
# 3

# Combined with field access:
echo '[{"name":"a","val":1},{"name":"b","val":2}]' | jq '.[].name'
# "a"
# "b"

# With -r for clean output:
echo '[{"name":"a"},{"name":"b"}]' | jq -r '.[].name'
# a
# b
```

## Pipes and Composition

jq has its own pipe operator `|` (inside the jq expression):

```bash
echo '{"users": [{"name": "Alice", "age": 30}, {"name": "Bob", "age": 25}]}' | \
  jq '.users[] | .name'
# "Alice"
# "Bob"

# Multiple pipes:
echo '{"data": {"items": [{"id": 1}, {"id": 2}]}}' | \
  jq '.data.items[] | .id'
# 1
# 2
```

## Filtering with select()

```bash
# Filter array elements:
pods='[
  {"name": "web-1", "status": "Running", "restarts": 0},
  {"name": "web-2", "status": "CrashLoopBackOff", "restarts": 15},
  {"name": "db-1", "status": "Running", "restarts": 0}
]'

# Select by condition:
echo "$pods" | jq '.[] | select(.status == "Running")'
# Shows web-1 and db-1 objects

# Select with regex:
echo "$pods" | jq '.[] | select(.name | test("web"))'
# Shows web-1 and web-2

# Select by numeric condition:
echo "$pods" | jq '.[] | select(.restarts > 0) | .name'
# "web-2"

# Combine conditions:
echo "$pods" | jq '.[] | select(.status == "Running" and .restarts == 0) | .name'
```

## Constructing Output

### Build Objects

```bash
echo '{"first": "Alice", "last": "Smith", "age": 30, "dept": "eng"}' | \
  jq '{name: (.first + " " + .last), department: .dept}'
# {"name": "Alice Smith", "department": "eng"}
```

### Build Arrays

```bash
echo '{"users": [{"name": "Alice", "role": "admin"}, {"name": "Bob", "role": "user"}]}' | \
  jq '[.users[] | .name]'
# ["Alice", "Bob"]

# With transformation:
echo '{"users": [{"name": "Alice"}, {"name": "Bob"}]}' | \
  jq '[.users[] | {user: .name, greeting: ("Hello, " + .name)}]'
```

### String Interpolation

```bash
echo '{"name": "Alice", "age": 30}' | jq -r '"\(.name) is \(.age) years old"'
# Alice is 30 years old

# In arrays:
echo '[{"name":"a","port":80},{"name":"b","port":443}]' | \
  jq -r '.[] | "\(.name):\(.port)"'
# a:80
# b:443
```

## Built-in Functions

```bash
# Length:
echo '[1,2,3]' | jq 'length'      # 3
echo '"hello"' | jq 'length'      # 5
echo '{"a":1,"b":2}' | jq 'length'  # 2 (number of keys)

# Keys and values:
echo '{"b":2,"a":1}' | jq 'keys'    # ["a","b"] (sorted)
echo '{"a":1,"b":2}' | jq 'values'  # [1,2]

# Map (transform each element):
echo '[1,2,3]' | jq 'map(. * 2)'    # [2,4,6]
echo '[{"n":"a","v":1},{"n":"b","v":2}]' | jq 'map(.v)'  # [1,2]

# Sort:
echo '[3,1,2]' | jq 'sort'          # [1,2,3]
echo '[{"name":"Bob"},{"name":"Alice"}]' | jq 'sort_by(.name)'

# Group:
echo '[{"t":"a","v":1},{"t":"b","v":2},{"t":"a","v":3}]' | \
  jq 'group_by(.t) | map({type: .[0].t, count: length})'

# Unique:
echo '[1,2,1,3,2]' | jq 'unique'    # [1,2,3]

# Flatten:
echo '[[1,2],[3,[4,5]]]' | jq 'flatten'  # [1,2,3,4,5]

# Add (sum arrays, merge objects):
echo '[1,2,3]' | jq 'add'           # 6
echo '[{"a":1},{"b":2}]' | jq 'add' # {"a":1,"b":2}

# Type checking:
echo '42' | jq 'type'               # "number"
echo '"hi"' | jq 'type'             # "string"
```

## Real-World Examples

### Docker

```bash
# List container names and statuses:
docker ps -a --format json | jq -r '[.Names, .Status] | @tsv'

# Inspect — get IP address:
docker inspect mycontainer | jq -r '.[0].NetworkSettings.IPAddress'

# Image sizes:
docker images --format json | jq -r '[.Repository, .Size] | @tsv' | sort -k2 -h
```

### Kubernetes (kubectl)

```bash
# Pod names in a namespace:
kubectl get pods -o json | jq -r '.items[].metadata.name'

# Pods not running:
kubectl get pods -o json | jq -r '
  .items[] | select(.status.phase != "Running") | 
  "\(.metadata.name): \(.status.phase)"
'

# Container images used:
kubectl get pods -o json | jq -r '
  [.items[].spec.containers[].image] | unique | .[]
'

# Resource requests/limits:
kubectl get pods -o json | jq -r '
  .items[] | .metadata.name as $pod |
  .spec.containers[] | 
  "\($pod)/\(.name): cpu=\(.resources.requests.cpu // "none") mem=\(.resources.requests.memory // "none")"
'
```

### GitHub API

```bash
# List repos:
curl -s "https://api.github.com/users/USERNAME/repos" | \
  jq -r '.[] | "\(.name)\t\(.stargazers_count)⭐\t\(.language // "unknown")"' | \
  sort -t$'\t' -k2 -rn

# Latest release:
curl -s "https://api.github.com/repos/OWNER/REPO/releases/latest" | \
  jq -r '.tag_name'
```

### Config File Processing

```bash
# Read a value from a JSON config:
DB_HOST=$(jq -r '.database.host' config.json)
DB_PORT=$(jq -r '.database.port' config.json)

# Modify a value:
jq '.database.port = 5433' config.json > config_new.json

# Add a field:
jq '.version = "2.0"' config.json | sponge config.json  # sponge from moreutils

# Merge two configs (second overrides first):
jq -s '.[0] * .[1]' defaults.json overrides.json
```

## jq with Shell Variables

```bash
# WRONG — unquoted variable in jq expression:
name="Alice"
echo '{}' | jq ".name = $name"    # Syntax error or injection risk!

# RIGHT — use --arg:
echo '{}' | jq --arg n "$name" '.name = $n'
# {"name": "Alice"}

# For numbers — use --argjson:
echo '{}' | jq --argjson port 8080 '.port = $port'
# {"port": 8080}   (number, not string)

# For raw input from a file:
jq --rawfile cert server.pem '.tls_cert = $cert' config.json

# Multiple arguments:
jq --arg host "$HOST" --argjson port "$PORT" \
   '.server = {host: $host, port: $port}' config.json
```

## Output Formats

```bash
# Tab-separated:
echo '[{"a":1,"b":2},{"a":3,"b":4}]' | jq -r '.[] | [.a, .b] | @tsv'
# 1	2
# 3	4

# CSV:
echo '[{"a":1,"b":2}]' | jq -r '.[] | [.a, .b] | @csv'
# 1,2

# URI encoding:
echo '"hello world"' | jq -r '@uri'
# hello%20world

# Base64:
echo '"hello"' | jq -r '@base64'
# aGVsbG8=

# Compact (single line):
echo '{"a": 1, "b": 2}' | jq -c '.'
# {"a":1,"b":2}
```

## Common Footguns

### 1. Forgetting `-r` for String Output

```bash
name=$(echo '{"name":"Alice"}' | jq '.name')
echo "$name"    # "Alice" — WITH QUOTES
# Use jq -r for raw strings:
name=$(echo '{"name":"Alice"}' | jq -r '.name')
```

### 2. null vs "null"

```bash
echo '{"a": null}' | jq '.a'        # null (JSON null)
echo '{"a": null}' | jq -r '.a'     # null (the STRING "null" — confusing!)

# Check for null properly:
echo '{"a": null}' | jq '.a // "default"'   # "default"
echo '{"a": null}' | jq 'if .a then .a else "missing" end'
```

### 3. Shell Quoting with jq

```bash
# Single quotes for jq expression, double quotes for shell:
jq '.items[] | select(.name == "web")' data.json    # Correct

# If the jq expression needs single quotes:
jq '.items[] | select(.name == "it'\''s")' data.json    # Escaped single quote
```

## Exercise

1. Fetch a JSON API (e.g., `curl -s https://jsonplaceholder.typicode.com/users`) and extract names and emails in "name <email>" format.

2. Write a jq filter that takes a list of objects and groups them by a field, outputting the count per group.

3. Use `kubectl get pods -o json` (or mock JSON) to find all pods with restart count > 0, printing the pod name, container name, and restart count.

4. Write a jq command that merges two JSON config files, with the second file's values taking priority.

---

Next: [Building Real Pipelines](03-real-pipelines.md)
