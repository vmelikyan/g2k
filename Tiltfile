k8s_yaml([
    'k8s/redpanda.yaml',
    'k8s/g2krelay.yaml',
    'k8s/g2krepeater.yaml',
    'k8s/redpanda-console/deployment.yaml',
    'k8s/redpanda-console/service.yaml',
])

k8s_resource(
    'redpanda-console',
    port_forwards=8080,
)

docker_build('g2krelay:dev', './g2krelay', dockerfile='./g2krelay/Dockerfile')
k8s_resource('g2krelay', port_forwards=5050)

docker_build('g2krepeater:dev', './g2krepeater', dockerfile='./g2krepeater/Dockerfile')
k8s_resource('g2krepeater')

watch_files = ['.g2krelay', '.g2krepeater']
for path in watch_files:
    local_resource(
        name='watch_' + path,
        cmd='true',
        deps=[path],
        resource_deps=['g2krelay', 'g2krepeater']
    )

