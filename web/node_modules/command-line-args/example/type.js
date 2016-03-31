var fs = require('fs')

function FileDetails (filename) {
  if (!(this instanceof FileDetails)) return new FileDetails(filename)
  this.filename = filename
  this.exists = fs.existsSync(filename)
}

module.exports = [
  { name: 'file', type: FileDetails },
  { name: 'depth', type: Number }
]
