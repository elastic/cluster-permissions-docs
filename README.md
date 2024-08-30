# Cluster permissions documentation generator

This is a very primitive, one-shot, project to convert a file that contains `ClusterRole` into a markdown table.

## Usage

```
Usage of cluster_permissions_doc_generator:
  -file string
        path to a file which contains a ClusterRole rules. Only one ClusterRole is supported.
```

## YAML Comments

It supports comments which precede a rule, for example:

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: my-role
rules:
# <ul><li>`Namespaces` are watched because...</li>
# <li>`Nodes` are watched because...</li>
# <li>`Endpoints` are watched because...</li></ul>
- apiGroups:
    - ""
  resources:
    - endpoints
    - namespaces
    - nodes
  verbs:
    - get
    - list
    - watch
```

This would be converted into the following:

```markdown
| API Groups  | Resources | Non Resource URLs | Resource Names | Verbs | Comment |
| ----------- | --------- | ----- | -------------- | ----------------- | ------- |
||<ul><li>endpoints</li><li>namespaces</li><li>nodes</li></ul>|none|none|<ul><li>get</li><li>list</li><li>watch</li></ul>|<ul><li>`Namespaces` are watched because...</li> <li>`Nodes` are watched because...</li> <li>`Endpoints` are watched because...</li></ul>|
```

And this would be displayed as:

| API Groups  | Resources | Non Resource URLs | Resource Names | Verbs | Comment |
| ----------- | --------- | ----- | -------------- | ----------------- | ------- |
||<ul><li>endpoints</li><li>namespaces</li><li>nodes</li></ul>|none|none|<ul><li>get</li><li>list</li><li>watch</li></ul>|<ul><li>`Namespaces` are watched because...</li> <li>`Nodes` are watched because...</li> <li>`Endpoints` are watched because...</li></ul>|

## Limitations

It has several strong limitations:
* Only one file can be provided.
* File must contain only one cluster role.
* Result is only generated to standard output.

