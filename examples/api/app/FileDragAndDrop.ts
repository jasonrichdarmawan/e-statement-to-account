const dropZone = <HTMLDivElement>document.querySelector("div[id=DropZone]");

if (dropZone == null) {
  throw Error("can't find element with selector: div[id=DropZone]");
}

var files: File[] = []

function renderFiles(files: File[]) {
  const fileListElement = <HTMLDivElement>document.querySelector("div[id=file-list]")
  if (fileListElement == null) {
    throw Error("cant find element with selector: div[id=file-list]")
  }

  if (files.length == 0) {
    fileListElement.setAttribute("hidden", "")
    return
  }

  if (fileListElement.hasAttribute("hidden")) {
    fileListElement.removeAttribute("hidden")
  }
  fileListElement.innerHTML = ""

  function buttonListener(e: MouseEvent) {
    const i = (<HTMLButtonElement>e.target).id
    files.splice(Number(i), 1)
    renderFiles(files)
  }

  for (let i = 0; i < files.length; i++) {
    const file = files[i]

    const fileElement = document.createElement("div")
    fileElement.className = "Box-row d-flex"

    const span1 = document.createElement("span")
    span1.className = "mr-2"
    const fileIcon = document.createElementNS("http://www.w3.org/2000/svg", "svg")
    fileIcon.setAttribute("data-src", "images/file.svg")
    fileIcon.setAttribute("class", "octicon")
    renderIcon(fileIcon)
    span1.appendChild(fileIcon)
    fileElement.appendChild(span1)

    const span2 = document.createElement("span")
    span2.className = "flex-auto css-truncate"
    span2.innerHTML = file.name
    fileElement.appendChild(span2)

    const button = document.createElement("button")
    button.setAttribute("type", "submit")
    button.className = "Link--primary btn-link"
    const xIcon = document.createElementNS("http://www.w3.org/2000/svg", "svg")
    xIcon.setAttribute("data-src", "images/x.svg")
    xIcon.setAttribute("class", "octicon")
    xIcon.setAttribute("id", i.toString())
    renderIcon(xIcon)
    button.appendChild(xIcon)
    fileElement.appendChild(button)
    button.addEventListener("click", buttonListener)

    fileListElement.appendChild(fileElement)
  }
}

function pushFiles(file: File): string | null {
  for (let i = 0; i < files.length; i++) {
    if (files[i].name === file.name) {
      return "file exist";
    }
  }
  files.push(file)
  return null;
}

dropZone.addEventListener("drop", function (ev) {
  // Prevent file from being opened
  ev.preventDefault();

  if (ev.dataTransfer == null) {
    throw Error("DropZone callback dataTransfer is null");
  }

  // Use DataTransferItemList interface to access the file(s)
  const items = ev.dataTransfer.items
  for (let i = 0; i < items.length; i++) {
    var item = items[i].webkitGetAsEntry()
    if (item == null) {
      throw Error("func webkitGetAsEntry return null");
    }
    scanFiles(item);
  }

  function scanFiles(item: FileSystemEntry) {
    if (item.isDirectory) {
      const directoryReader = (<FileSystemDirectoryEntry>item).createReader();
      directoryReader.readEntries((entries) => {
        entries.forEach((entry) => {
          scanFiles(entry);
        });
      });
    } else if (item.isFile) {
      (<FileSystemFileEntry>item).file(function (file) {
        if (file.type != "application/pdf") {
          return
        }

        pushFiles(file)

        renderFiles(files)
      });
    }
  }
});

dropZone.addEventListener("dragover", function (ev) {
  // Prevent file from being opened.
  ev.preventDefault();
});

const inputFilesElement = <HTMLInputElement>document.querySelector("input[type='file'][id='files']")

if (inputFilesElement == null) {
  throw Error("can't find element with selector: input[type='file'][id='files']")
}

inputFilesElement.addEventListener("change", (ev) => {
  const fileList = (<HTMLInputElement>ev.target).files
  if (fileList == null) {
    throw Error("FileList object is null")
  }
  console.log(fileList.length)
  for (let i = 0; i < fileList.length; i++) {
    let err = pushFiles(fileList[i])
    if (err != null) {
      continue;
    }
    renderFiles(files);
  }
  (<HTMLInputElement>ev.target).value = "";
})

const buttonUploadElement = <HTMLButtonElement>document.querySelector("button[id='/parser']")

buttonUploadElement.addEventListener("click", () => {
  if (files.length == 0) {
    throw Error("Files can't be empty")
  }

  const parseToTableElement = <HTMLInputElement>document.querySelector("input[type='checkbox'][name='ParseToTable']")
  if (parseToTableElement == null) {
    throw Error("can't find element with selector: input[type='checkbox'][name='ParseToTable']")
  }

  const groupByAccountElement = <HTMLInputElement>document.querySelector("input[type='checkbox'][name='GroupByAccount']")
  if (groupByAccountElement == null) {
    throw Error("can't find element with selector: input[type='checkbox'][name='GroupByAccount']")
  }

  const summaryByAccountElement = <HTMLInputElement>document.querySelector("input[type='checkbox'][name='SummaryByAccount']")
  if (summaryByAccountElement == null) {
    throw Error("can't find element with selector: input[type='checkbox'][name='SummaryByAccount']")
  }

  if (!(parseToTableElement.checked || groupByAccountElement.checked || summaryByAccountElement.checked)) {
    throw Error("OPTIONS can't be empty")
  }

  const formData = new FormData();
  formData.append("ParseToTable", parseToTableElement.checked ? "true" : "")
  formData.append("GroupByAccount",
    groupByAccountElement.checked ? "true" :
      summaryByAccountElement.checked ? "true" : "")
  files.forEach((file) => {
    formData.append("Files", file)
  })

  fetch("/parser", {
    method: "POST",
    body: formData
  }).then((response) => response.json())
    .then((json) => {
      CleanAllRender()

      if (parseToTableElement.checked) {
        RenderTransactions(json)
      }
      if (groupByAccountElement.checked) {
        RenderGroupedByAccount(json)
      }
      if (summaryByAccountElement.checked) {
        RenderAccountSummary(json)
      }
    })
})

const cancelElement = <HTMLLabelElement> document.querySelector("label[id='cancelparser']")

if (cancelElement == null) {
  throw Error("element can't be null")
}

cancelElement.addEventListener("click", () => {
  files = []
  renderFiles(files)

  CleanAllRender()
})