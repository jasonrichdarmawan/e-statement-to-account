var CheckBox = <HTMLInputElement> document.querySelector('input[name=darkmode]')

if (!(CheckBox instanceof HTMLInputElement)) {
    throw Error("var CheckBox is not instance of HTMLInputElement")
}

CheckBox.addEventListener('change', function() {
    if (this.checked) {
        document.documentElement.setAttribute('data-theme', 'dark')
    } else {
        document.documentElement.removeAttribute('data-theme')
    }
})