export enum EnvType {
  K8S = 'kubernetes',
  DOCKER = 'docker',
}

export enum EnvFormData {
  TIP_CONFIG = `
    apiVersion: v1
    kind: ConfigMap
    metadata:
    name: example
    data:
    key: value
  `,
}
