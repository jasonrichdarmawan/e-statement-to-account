var FormEStatementToAccount = <HTMLFormElement> document.querySelector("form[action='/e-statement-to-account']")

if (!(FormEStatementToAccount instanceof HTMLFormElement)) {
    throw new Error("var DropZone is not instanceof HTMLDivElement")
}

// TODO: this is just a mock-up
FormEStatementToAccount.addEventListener('drop', function(ev) {
    // Prevent file from being opened
    ev.preventDefault()

    if (ev.dataTransfer == null) {
        throw new Error("DropZone callback dataTransfer is null")
    }

    if (ev.dataTransfer.items) {
        // Use DataTransferItemList interface to access the file(s)
        [...ev.dataTransfer.items].forEach((item, i) => {
            // If dropped items aren't files, reject them
            if (item.kind === 'file') {
              const file = item.getAsFile();
              console.log(`… file[${i}].name = ${file?.name}`);
            }
          });
    } else {
        // Use DataTransfer interface to access the file(s)
        [...ev.dataTransfer.files].forEach((file, i) => {
            console.log(`… file[${i}].name = ${file.name}`);
        });
    }
})

FormEStatementToAccount.addEventListener('dragover', function(ev) {
    // Prevent file from being opened.
    ev.preventDefault()
})