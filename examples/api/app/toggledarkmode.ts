var checkBox = <HTMLInputElement>document.querySelector('input[name=darkmode]')

if (!(checkBox instanceof HTMLInputElement)) {
    throw Error("var CheckBox is not instance of HTMLInputElement")
}

checkBox.addEventListener('change', function () {
    if (this.checked) {
        document.documentElement.setAttribute('data-theme', 'dark')
    } else {
        document.documentElement.removeAttribute('data-theme')
    }
})