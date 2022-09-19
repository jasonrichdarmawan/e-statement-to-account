const accordionElements = document.getElementsByClassName("accordion")

for (let i = 0; i < accordionElements.length; i++) {
    (accordionElements[i]).addEventListener("click", function(e) {
        const accordionElement = <HTMLDivElement>e.target
        const panelElement = (<HTMLDivElement>accordionElement.nextElementSibling);
        if (panelElement == null) {
            throw Error("can't find the next element")
        }
        panelElement.classList.toggle("panel")
    })
}