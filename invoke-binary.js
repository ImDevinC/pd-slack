const childProcess = require('child_process')
const os = require('os')
const process = require('process')

function chooseBinary() {
  const binaries = {
    linux: {
      x64: 'pd-slack-linux-amd64',
      arm64: 'pd-slack-linux-arm64',
    },
  }

  const binary = (binaries[os.platform()] || {})[os.arch()]
  if (!binary) {
    console.error(
      `pd-slack: no prebuilt binary for platform ${os.platform()} (${os.arch()}); ` +
        'supported platforms are linux/amd64 and linux/arm64'
    )
    process.exit(1)
  }
  return `${__dirname}/dist/${binary}`
}

function main() {
  const binary = chooseBinary()
  const result = childProcess.spawnSync(binary, { stdio: 'inherit' })
  if (typeof result.status === 'number') {
    process.exit(result.status)
  }
  if (result.error) {
    console.error(`pd-slack: failed to run ${binary}: ${result.error.message}`)
  }
  process.exit(1)
}

if (require.main === module) {
  main()
}