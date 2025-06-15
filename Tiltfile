docker_build('localhost:5000/runway', '.',
  dockerfile='Dockerfile.development',
  live_update=[
    sync('.', '/app'),
  ]
)

k8s_yaml([
  'k8s/namespace.yaml',
  'k8s/registry.yaml',
  'k8s/postgres.yaml',
  'k8s/deployment.yaml',
])

allow_k8s_contexts('default')
k8s_resource('runway', port_forwards=[8080, 7331])
