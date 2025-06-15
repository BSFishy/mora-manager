docker_build('localhost:5050/runway', '.',
  dockerfile='Dockerfile.development',
  live_update=[
    sync('.', '/app'),
    run('find /tmp -maxdepth 1 -name "go-*" | xargs -r rm -r'),
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
