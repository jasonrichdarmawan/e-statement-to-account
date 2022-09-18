var DropZone = <HTMLDivElement>document.querySelector("div[id=DropZone]");

if (!(DropZone instanceof HTMLDivElement)) {
  throw Error("var DropZone is not instanceof HTMLDivElement");
}

var files: File[] = []

// TODO: this is just a mock-up
DropZone.addEventListener("drop", function (ev) {
  // Prevent file from being opened
  ev.preventDefault();

  if (ev.dataTransfer == null) {
    throw Error("DropZone callback dataTransfer is null");
  }

  // Use DataTransferItemList interface to access the file(s)
  let items = ev.dataTransfer.items
  for (let i = 0; i < items.length; i++) {
    var item = items[i].webkitGetAsEntry()
    if (item == null) {
        throw Error("func webkitGetAsEntry return null");
      }
      scanFiles(item);
  }

  function scanFiles(item: FileSystemEntry) {
    if (item.isDirectory) {
      let directoryReader = (<FileSystemDirectoryEntry>item).createReader();
      directoryReader.readEntries((entries) => {
        entries.forEach((entry) => {
          scanFiles(entry);
        });
      });
    } else if (item.isFile) {
      (<FileSystemFileEntry>item).file(function (file) {
        let fileListElement = <HTMLDivElement> document.querySelector("div[id=file-list]")
        if (fileListElement == null) {
            throw Error("let fileList is null")
        }
        if (fileListElement.hasAttribute("hidden")) {
            fileListElement.removeAttribute("hidden")
        }
        files.push(file)
        let fileElement = document.createElement("div")
        fileElement.className = "Box-row d-flex"

        let span1 = document.createElement("span")
        span1.className = "mr-2"
        let fileIcon = document.createElementNS("http://www.w3.org/2000/svg", "svg")
        fileIcon.setAttribute("data-src", "images/file.svg")
        fileIcon.setAttribute("class", "octicon")
        renderIcon(fileIcon)
        span1.appendChild(fileIcon)
        fileElement.appendChild(span1)

        let span2 = document.createElement("span")
        span2.className = "flex-auto css-truncate"
        span2.innerHTML = file.name
        fileElement.appendChild(span2)

        let button = document.createElement("button")
        button.setAttribute("type", "submit")
        button.className = "Link--primary btn-link"
        let xIcon = document.createElementNS("http://www.w3.org/2000/svg", "svg")
        xIcon.setAttribute("data-src", "images/x.svg")
        xIcon.setAttribute("class", "octicon")
        renderIcon(xIcon)
        button.appendChild(xIcon)
        fileElement.appendChild(button)

        fileListElement.appendChild(fileElement)
      });
    }
  }
});

DropZone.addEventListener("dragover", function (ev) {
  // Prevent file from being opened.
  ev.preventDefault();
});