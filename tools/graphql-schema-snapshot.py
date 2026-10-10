"""Capture schema metadata only; never write authentication or endpoint data."""
import argparse
from datetime import datetime, timezone
import json
from pathlib import Path
import urllib.request


def reference(value):
    if value["kind"] == "NON_NULL":
        return reference(value["ofType"]) + "!"
    if value["kind"] == "LIST":
        return "[" + reference(value["ofType"]) + "]"
    return value["name"]


def input_value(value):
    result = value["name"] + ": " + reference(value["type"])
    if value.get("defaultValue") is not None:
        result += " = " + value["defaultValue"]
    return result


def render(schema):
    roots = [f"  {kind}: {schema[kind + 'Type']['name']}" for kind in
             ("query", "mutation", "subscription") if schema.get(kind + "Type")]
    output = [f"# Live North America schema captured {datetime.now(timezone.utc).date()}; metadata only.",
              "schema {\n" + "\n".join(roots) + "\n}"]
    for value in sorted(schema["types"], key=lambda item: item["name"]):
        name, kind = value["name"], value["kind"]
        if name.startswith("__") or name in ("Int", "Float", "String", "Boolean", "ID"):
            continue
        if kind == "SCALAR":
            output.append("scalar " + name)
        elif kind == "UNION":
            output.append("union " + name + " = " + " | ".join(item["name"] for item in value["possibleTypes"]))
        else:
            keyword = {"OBJECT": "type", "INTERFACE": "interface", "INPUT_OBJECT": "input", "ENUM": "enum"}[kind]
            heading = keyword + " " + name
            if value.get("interfaces"):
                heading += " implements " + " & ".join(item["name"] for item in value["interfaces"])
            fields = []
            for field in value.get("fields") or []:
                args = ", ".join(input_value(arg) for arg in field["args"])
                fields.append(field["name"] + ("(" + args + ")" if args else "") + ": " + reference(field["type"]))
            fields += [input_value(field) for field in value.get("inputFields") or []]
            fields += [field["name"] for field in value.get("enumValues") or []]
            output.append(heading + " {\n" + "\n".join("  " + field for field in fields) + "\n}")
    return "\n\n".join(output) + "\n"


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--credentials", required=True, type=Path)
    parser.add_argument("--output", required=True, type=Path)
    args = parser.parse_args()
    credentials = json.loads(args.credentials.read_text(encoding="utf-8"))
    ref = "kind name"
    for _ in range(7):
        ref = "kind name ofType { " + ref + " }"
    field_type = "type { " + ref + " }"
    inputs = "name defaultValue " + field_type
    query = "query SchemaSnapshot { __schema { queryType { name } mutationType { name } subscriptionType { name } types { kind name fields(includeDeprecated:true) { name args { " + inputs + " } " + field_type + " } inputFields { " + inputs + " } enumValues(includeDeprecated:true) { name } interfaces { name } possibleTypes { name } } } }"
    request = urllib.request.Request("https://alexa.amazon.com/nexus/v1/graphql", data=json.dumps({"query": query, "variables": {}}).encode(), headers={"Authorization": "Bearer " + credentials["accessToken"], "Content-Type": "application/json"})
    response = json.loads(urllib.request.urlopen(request, timeout=45).read())
    if response.get("errors"):
        raise SystemExit("Schema introspection failed; no schema written")
    schema = response["data"]["__schema"]
    args.output.write_text(render(schema), encoding="utf-8")
    print(f"Wrote schema metadata for {len(schema['types'])} types")


if __name__ == "__main__":
    main()
