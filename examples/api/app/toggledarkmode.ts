var checkbox = <HTMLInputElement> document.querySelector('input[name=darkmode]')

if (!(checkbox instanceof HTMLInputElement)) {
    alert("checkbox is null")
}

checkbox.addEventListener('change', function() {
    if (this.checked) {
        document.documentElement.setAttribute('data-theme', 'dark')
    } else {
        document.documentElement.removeAttribute('data-theme')
    }
})