// Copyright 2022 Nho Luong DevOps.
// SPDX-License-Identifier: Apache-2.0

package krusty_test

import (
	"testing"

	kusttest_test "sigs.k8s.io/kustomize/api/testutils/kusttest"
)

func TestIssue3489Simplified(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	th.WriteK(".", `
namespace: kube-system
resources:
- aa
- bb
`)
	th.WriteK("aa", `
resources:
- ../base
`)
	th.WriteK("bb", `
resources:
- ../base
nameSuffix: -private
`)
	th.WriteK("base", `
resources:
- deployment.yaml
- serviceAccount.yaml
`)
	th.WriteF("base/deployment.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myDep
spec:
  template:
    spec:
      serviceAccountName: mySvcAcct
      containers:
      - name: whatever
        image: registry.k8s.io/governmentCheese
`)
	th.WriteF("base/serviceAccount.yaml", `
apiVersion: v1
kind: ServiceAccount
metadata:
  name: mySvcAcct
`)
	m := th.Run(".", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(m, `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myDep
  namespace: kube-system
spec:
  template:
    spec:
      containers:
      - image: registry.k8s.io/governmentCheese
        name: whatever
      serviceAccountName: mySvcAcct
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: mySvcAcct
  namespace: kube-system
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myDep-private
  namespace: kube-system
spec:
  template:
    spec:
      containers:
      - image: registry.k8s.io/governmentCheese
        name: whatever
      serviceAccountName: mySvcAcct-private
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: mySvcAcct-private
  namespace: kube-system
`)
}

func TestIssue3489(t *testing.T) {
	const assets = `{
	"tenantId": "XXXXX-XXXXXX-XXXXX-XXXXXX-XXXXXX",
	"subscriptionId": "XXXXX-XXXXXX-XXXXX-XXXXXX-XXXXXX",
	"resourceGroup": "DNS-EUW-XXX-RG",
	"useManagedIdentityExtension": true,
	"userAssignedIdentityID": "XXXXX-XXXXXX-XXXXX-XXXXXX-XXXXXX"
}
`
	th := kusttest_test.MakeHarness(t)
	th.WriteK(".", `
namespace: kube-system
resources:
- external-dns
- external-dns-private
`)
	th.WriteK("external-dns", `
resources:
- ../base
commonLabels:
  app: external-dns
  instance: public
images:
- name: registry.k8s.io/external-dns/external-dns
  newName: xxx.azurecr.io/external-dns
  newTag: v0.7.4_sylr.1
- name: quay.io/sylr/external-dns
  newName: xxx.azurecr.io/external-dns
  newTag: v0.7.4_sylr.1
secretGenerator:
- name: azure-config-file
  behavior: replace
  files:
  - assets/azure.json
patches:
- target:
    group: apps
    version: v1
    kind: Deployment
    name: external-dns
  patch: |-
    - op: replace
      path: /spec/template/spec/containers/0/args
      value:
      - --txt-owner-id="aks"
      - --txt-prefix=external-dns-
      - --source=service
      - --provider=azure
      - --registry=txt
      - --domain-filter=dev.company.com
`)

	th.WriteF("external-dns/assets/azure.json", assets)
	th.WriteK("external-dns-private", `
resources:
- ../base
nameSuffix: -private
commonLabels:
  app: external-dns
  instance: private
images:
- name: registry.k8s.io/external-dns/external-dns
  newName: xxx.azurecr.io/external-dns
  newTag: v0.7.4_sylr.1
- name: quay.io/sylr/external-dns
  newName: xxx.azurecr.io/external-dns
  newTag: v0.7.4_sylr.1
secretGenerator:
- name: azure-config-file
  behavior: replace
  files:
  - assets/azure.json
patches:
- target:
    group: apps
    version: v1
    kind: Deployment
    name: external-dns
  patch: |-
    - op: replace
      path: /spec/template/spec/containers/0/args
      value:
      - --txt-owner-id="aks"
      - --txt-prefix=external-dns-private-
      - --source=service
      - --provider=azure-private-dns
      - --registry=txt
      - --domain-filter=static.company.az
`)
	th.WriteF("external-dns-private/assets/azure.json", assets)
	th.WriteK("base", `
resources:
- clusterrole.yaml
- clusterrolebinding.yaml
- deployment.yaml
- serviceaccount.yaml
commonLabels:
  app: external-dns
  instance: public
images:
- name: registry.k8s.io/external-dns/external-dns
  newName: quay.io/sylr/external-dns
  newTag: v0.7.4-73-g00a9a0c7
secretGenerator:
- name: azure-config-file
  files:
  - assets/azure.json
`)
	th.WriteF("base/assets/azure.json", assets)
	th.WriteF("base/clusterrolebinding.yaml", `
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: external-dns-viewer
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: external-dns
subjects:
- kind: ServiceAccount
  name: external-dns
`)
	th.WriteF("base/clusterrole.yaml", `
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: external-dns
rules:
- apiGroups: ['']
  resources: ['endpoints', 'pods', 'services', 'nodes']
  verbs: ['get', 'watch', 'list']
- apiGroups: ['extensions', 'networking.k8s.io']
  resources: ['ingresses']
  verbs: ['get', 'watch', 'list']
`)
	th.WriteF("base/deployment.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: external-dns
spec:
  strategy:
    type: Recreate
  selector:
    matchLabels: {}
  template:
    metadata: {}
    spec:
      serviceAccountName: external-dns
      containers:
      - name: external-dns
        image: registry.k8s.io/external-dns/external-dns
        args:
        - --domain-filter=""
        - --txt-owner-id=""
        - --txt-prefix=external-dns-
        - --source=service
        - --provider=azure
        - --registry=txt
        resources: {}
        volumeMounts:
        - name: azure-config-file
          mountPath: /etc/kubernetes
          readOnly: true
      volumes:
      - name: azure-config-file
        secret:
          secretName: azure-config-file
`)
	th.WriteF("base/serviceaccount.yaml", `
apiVersion: v1
kind: ServiceAccount
metadata:
  name: external-dns
`)
	m := th.Run(".", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(
		m, `
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  labels:
    app: external-dns
    instance: public
  name: external-dns
rules:
- apiGroups:
  - ""
  resources:
  - endpoints
  - pods
  - services
  - nodes
  verbs:
  - get
  - watch
  - list
- apiGroups:
  - extensions
  - networking.k8s.io
  resources:
  - ingresses
  verbs:
  - get
  - watch
  - list
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  labels:
    app: external-dns
    instance: public
  name: external-dns-viewer
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: external-dns
subjects:
- kind: ServiceAccount
  name: external-dns
  namespace: kube-system
---
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: external-dns
    instance: public
  name: external-dns
  namespace: kube-system
spec:
  selector:
    matchLabels:
      app: external-dns
      instance: public
  strategy:
    type: Recreate
  template:
    metadata:
      labels:
        app: external-dns
        instance: public
    spec:
      containers:
      - args:
        - --txt-owner-id="aks"
        - --txt-prefix=external-dns-
        - --source=service
        - --provider=azure
        - --registry=txt
        - --domain-filter=dev.company.com
        image: xxx.azurecr.io/external-dns:v0.7.4_sylr.1
        name: external-dns
        resources: {}
        volumeMounts:
        - mountPath: /etc/kubernetes
          name: azure-config-file
          readOnly: true
      serviceAccountName: external-dns
      volumes:
      - name: azure-config-file
        secret:
          secretName: azure-config-file-66cc4224mm
---
apiVersion: v1
kind: ServiceAccount
metadata:
  labels:
    app: external-dns
    instance: public
  name: external-dns
  namespace: kube-system
---
apiVersion: v1
data:
  azure.json: |
    ewoJInRlbmFudElkIjogIlhYWFhYLVhYWFhYWC1YWFhYWC1YWFhYWFgtWFhYWFhYIiwKCS
    JzdWJzY3JpcHRpb25JZCI6ICJYWFhYWC1YWFhYWFgtWFhYWFgtWFhYWFhYLVhYWFhYWCIs
    CgkicmVzb3VyY2VHcm91cCI6ICJETlMtRVVXLVhYWC1SRyIsCgkidXNlTWFuYWdlZElkZW
    50aXR5RXh0ZW5zaW9uIjogdHJ1ZSwKCSJ1c2VyQXNzaWduZWRJZGVudGl0eUlEIjogIlhY
    WFhYLVhYWFhYWC1YWFhYWC1YWFhYWFgtWFhYWFhYIgp9Cg==
kind: Secret
metadata:
  labels:
    app: external-dns
    instance: public
  name: azure-config-file-66cc4224mm
  namespace: kube-system
type: Opaque
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  labels:
    app: external-dns
    instance: private
  name: external-dns-private
rules:
- apiGroups:
  - ""
  resources:
  - endpoints
  - pods
  - services
  - nodes
  verbs:
  - get
  - watch
  - list
- apiGroups:
  - extensions
  - networking.k8s.io
  resources:
  - ingresses
  verbs:
  - get
  - watch
  - list
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  labels:
    app: external-dns
    instance: private
  name: external-dns-viewer-private
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: external-dns-private
subjects:
- kind: ServiceAccount
  name: external-dns-private
  namespace: kube-system
---
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: external-dns
    instance: private
  name: external-dns-private
  namespace: kube-system
spec:
  selector:
    matchLabels:
      app: external-dns
      instance: private
  strategy:
    type: Recreate
  template:
    metadata:
      labels:
        app: external-dns
        instance: private
    spec:
      containers:
      - args:
        - --txt-owner-id="aks"
        - --txt-prefix=external-dns-private-
        - --source=service
        - --provider=azure-private-dns
        - --registry=txt
        - --domain-filter=static.company.az
        image: xxx.azurecr.io/external-dns:v0.7.4_sylr.1
        name: external-dns
        resources: {}
        volumeMounts:
        - mountPath: /etc/kubernetes
          name: azure-config-file
          readOnly: true
      serviceAccountName: external-dns-private
      volumes:
      - name: azure-config-file
        secret:
          secretName: azure-config-file-private-66cc4224mm
---
apiVersion: v1
kind: ServiceAccount
metadata:
  labels:
    app: external-dns
    instance: private
  name: external-dns-private
  namespace: kube-system
---
apiVersion: v1
data:
  azure.json: |
    ewoJInRlbmFudElkIjogIlhYWFhYLVhYWFhYWC1YWFhYWC1YWFhYWFgtWFhYWFhYIiwKCS
    JzdWJzY3JpcHRpb25JZCI6ICJYWFhYWC1YWFhYWFgtWFhYWFgtWFhYWFhYLVhYWFhYWCIs
    CgkicmVzb3VyY2VHcm91cCI6ICJETlMtRVVXLVhYWC1SRyIsCgkidXNlTWFuYWdlZElkZW
    50aXR5RXh0ZW5zaW9uIjogdHJ1ZSwKCSJ1c2VyQXNzaWduZWRJZGVudGl0eUlEIjogIlhY
    WFhYLVhYWFhYWC1YWFhYWC1YWFhYWFgtWFhYWFhYIgp9Cg==
kind: Secret
metadata:
  labels:
    app: external-dns
    instance: private
  name: azure-config-file-private-66cc4224mm
  namespace: kube-system
type: Opaque
`)
}

func TestEmptyFieldSpecValue(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	th.WriteK(".", `
generators:
- generators.yaml
configurations:
- kustomizeconfig.yaml
`)
	th.WriteF("generators.yaml", `
apiVersion: builtin
kind: ConfigMapGenerator
metadata:
  name: secret-example
labels:
  app.kubernetes.io/name: secret-example
literals:
- this_is_a_secret_name=
`)
	th.WriteF("kustomizeconfig.yaml", `
nameReference:
- kind: Secret
  version: v1
  fieldSpecs:
  - path: data/this_is_a_secret_name
    kind: ConfigMap
`)
	m := th.Run(".", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(m, `
apiVersion: v1
data:
  this_is_a_secret_name: ""
kind: ConfigMap
metadata:
  name: secret-example-7hf4fh868h
`)
}

func TestUnrelatedNameReferenceReplacement_Issue4254_Issue3418(t *testing.T) {
	th := kusttest_test.MakeHarness(t)

	// The cluster-autoscaler lease name should not be changed.
	th.WriteF("role.yaml", `
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: cluster-autoscaler
rules:
- apiGroups: ["coordination.k8s.io"]
  resources: ["leases"]
  resourceNames: ["cluster-autoscaler"]
  verbs: ["get","update"]
`)

	th.WriteK(".", `
resources:
- role.yaml
configMapGenerator:
- name: cluster-autoscaler
  namespace: kube-system
  literals:
  - AWS_REGION="us-east-1"
`)
	// The resourceNames for the leases resource in the ClusterRole should NOT be
	// updated with the name suffix, because it's not targeting the generated
	// configmap. The value at rules[0].resourceNames[0] is currently incorrect.
	m := th.Run(".", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(m, `
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: cluster-autoscaler
rules:
- apiGroups:
  - coordination.k8s.io
  resourceNames:
  - cluster-autoscaler-h8mmcct52k
  resources:
  - leases
  verbs:
  - get
  - update
---
apiVersion: v1
data:
  AWS_REGION: us-east-1
kind: ConfigMap
metadata:
  name: cluster-autoscaler-h8mmcct52k
  namespace: kube-system
`)
}

func TestIssue4682_NameReferencesToSelfInAnnotations(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	th.WriteK(".", `
namespace: newNs
resources:
  - resources.yaml

nameSuffix: -updated

configurations:
  - kustomize-nameref.yaml 
`)
	th.WriteF("kustomize-nameref.yaml", `
nameReference:
  - kind: Namespace
    fieldSpecs:
      - path: data/theNamespace
        kind: ConfigMap
        version: v1
      - path: metadata/annotations/theNamespace
        kind: ConfigMap
        version: v1      
      - path: metadata/annotations/theNamespace
        kind: Namespace
        version: v1
  - kind: ConfigMap
    fieldSpecs:
      - path: data/theConfigMap
        kind: ConfigMap
        version: v1
      - path: metadata/annotations/theConfigMap
        kind: ConfigMap
        version: v1   
      - path: metadata/annotations/theConfigMap
        kind: Namespace
        version: v1
`)
	th.WriteF("resources.yaml", `
apiVersion: v1
kind: ConfigMap
metadata:
  annotations:
    theConfigMap: cm
    theNamespace: oldNs
  name: cm
  namespace: oldNs
data:
  theConfigMap: cm
  theNamespace: oldNs
---
apiVersion: v1
kind: Namespace
metadata:
  annotations:
    theConfigMap: cm
    theNamespace: oldNs
  name: oldNs
`)
	m := th.Run(".", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(m, `
apiVersion: v1
data:
  theConfigMap: cm-updated
  theNamespace: newNs
kind: ConfigMap
metadata:
  annotations:
    theConfigMap: cm-updated
    theNamespace: newNs
  name: cm-updated
  namespace: newNs
---
apiVersion: v1
kind: Namespace
metadata:
  annotations:
    theConfigMap: cm-updated
    theNamespace: newNs
  name: newNs
`)
}

func TestIssue4884_UseLocalConfigAsNameRefSource(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	th.WriteK(".", `
resources:
  - resources.yaml

namePrefix: prefix-

configurations:
  - kustomize-nameref.yaml
`)
	th.WriteF("kustomize-nameref.yaml", `
nameReference:
- kind: IngressHost
  fieldSpecs:
  - path: spec/rules/host
    kind: Ingress
  - path: spec/tls/hosts
    kind: Ingress
  - path: spec/template/spec/containers/env/value
    kind: Deployment
- kind: IngressSecret
  fieldSpecs:
  - path: spec/tls/secretName
    kind: Ingress
namePrefix:
- path: metadata/name
  kind: IngressHost
- path: metadata/name
  kind: IngressSecret

`)
	th.WriteF("resources.yaml", `
apiVersion: local/v1
kind: IngressHost
metadata:
  name: test.fakedomain.com
  namespace: test
  annotations:
    config.kubernetes.io/local-config: "true"
---
apiVersion: local/v1
kind: IngressSecret
metadata:
  name: test-secret
  namespace: test
  annotations:
    config.kubernetes.io/local-config: "true"
---
apiVersion: v1
kind: Ingress
metadata:
  name: test-ingress
  namespace: test
spec:
  rules:
  - host: test.fakedomain.com
  - host: do-not-touch.otherdomain.com
  tls:
  - hosts:
    - test.fakedomain.com
    secretName: test-secret
  - hosts:
    - do-not-touch.otherdomain.com
    secretname: do-not-touch
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-deployment
  namespace: test
spec:
  template:
    spec:
      containers:
      - name: tester
        env:
        - name: domain-name
          value: test.fakedomain.com
`)
	m := th.Run(".", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(m, `
apiVersion: v1
kind: Ingress
metadata:
  name: test-ingress
  namespace: test
spec:
  rules:
  - host: prefix-test.fakedomain.com
  - host: do-not-touch.otherdomain.com
  tls:
  - hosts:
    - prefix-test.fakedomain.com
    secretName: prefix-test-secret
  - hosts:
    - do-not-touch.otherdomain.com
    secretname: do-not-touch
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-deployment
  namespace: test
spec:
  template:
    spec:
      containers:
      - env:
        - name: domain-name
          value: prefix-test.fakedomain.com
        name: tester
`)
}

func TestBackReferenceAdmissionPolicy(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	th.WriteK(".", `
resources:
- admission.yaml

namePrefix: a-prefix-
`)
	th.WriteF("admission.yaml", `---
apiVersion: admissionregistration.k8s.io/v1beta1
kind: ValidatingAdmissionPolicy
metadata:
  name: sample-policy
spec:
  failurePolicy: Fail
  paramKind:
    apiVersion: apps/v1
    kind: Deployment
  matchConstraints:
    resourceRules:
    - apiGroups:
      - apps
      apiVersions:
      - v1
      operations:
      - CREATE
      - UPDATE
      resources:
      - deployments
  validations:
    - expression: "!object.metadata.name.startsWith('test-')"
      message: prefix 'test-' is not allowed
      reason: Invalid
---
apiVersion: admissionregistration.k8s.io/v1beta1
kind: ValidatingAdmissionPolicyBinding
metadata:
  name: sample-policy-binding
spec:
  policyName: sample-policy
  validationActions:
  - Deny
`)

	m := th.Run(".", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(m, `
apiVersion: admissionregistration.k8s.io/v1beta1
kind: ValidatingAdmissionPolicy
metadata:
  name: a-prefix-sample-policy
spec:
  failurePolicy: Fail
  matchConstraints:
    resourceRules:
    - apiGroups:
      - apps
      apiVersions:
      - v1
      operations:
      - CREATE
      - UPDATE
      resources:
      - deployments
  paramKind:
    apiVersion: apps/v1
    kind: Deployment
  validations:
  - expression: '!object.metadata.name.startsWith(''test-'')'
    message: prefix 'test-' is not allowed
    reason: Invalid
---
apiVersion: admissionregistration.k8s.io/v1beta1
kind: ValidatingAdmissionPolicyBinding
metadata:
  name: a-prefix-sample-policy-binding
spec:
  policyName: a-prefix-sample-policy
  validationActions:
  - Deny
`)
}
