/*
https://gist.github.com/codeguy/6684588#gistcomment-3243980
 */
function slugify(text) {
    return text
        .toString()                     // Cast to string
        .toLowerCase()                  // Convert the string to lowercase letters
        .normalize('NFD')       // The normalize() method returns the Unicode Normalization Form of a given string.
        .trim()                         // Remove whitespace from both sides of a string
        .replace(/\s+/g, '-')           // Replace spaces with -
        .replace(/[^\w\-]+/g, '')       // Remove all non-word chars
        .replace(/\-\-+/g, '-');        // Replace multiple - with single -
}

// Custom-Element my-element anlegen
class sdFont extends HTMLElement {

    // Festlegen, welche Attribute überwacht werden solle
    static get observedAttributes() {
        return ['face', 'size'];
    }

    constructor() {
        // Element wird angelegt

        // super muss als Erstes im constructor aufgerufen werden, super ruft den constructor der Elternklasse auf
        super();


        // Schatten-Dom anlegen
        // mode: 'open' : Vom Dokument aus ist der Zugriff auf das Schatten-Dom möglich.
        // mode: 'closed' : Der Zugriff auf das Schatten-Dom ist nicht möglich.
        this.shadow = this.attachShadow({mode: 'open'});

        // Element für Inhalt anlegen und ins Schatten-Dom einhängen
        const div = document.createElement('div');
        const style = document.createElement('style');
        this.shadow.appendChild(style);
        this.shadow.appendChild(div);

        // Weiterer Code
    }

    connectedCallback() {
        // Element wurde ins DOM eingehängt
    }

    disconnectedCallback () {
        // Element wurde entfernt
    }

    adoptedCallback() {
        // Element ist in ein anderes Dokument umgezogen
    }

    attributeChangedCallback(name, oldValue, newValue) {
        // Elementparameter wurden geändert
        // Achtung attributeChangedCallback wird vor connectedCallback aufgerufen

        let className = name+'-'+slugify(newValue);
        this.classList.add(className)
    }

}
customElements.define('sd-font', sdFont);

function initNetscape() {

    console.log("initNetscape(): "+window.location.href)
    // all elements which get class from attributes
    const classElements = ['font'];
    classElements.forEach((elemName) => {
        let elems = document.getElementsByTagName(elemName);
        Array.from(elems).forEach((elem) => {
            Array.from(elem.getAttributeNames()).forEach((attrName) => {
                let className = attrName.toLowerCase() + '-' + slugify(elem.getAttribute(attrName))
                elem.classList.add(className)
            });
        });
    });

    // all elements which must be replaced with sd-<name> custom web elements
    const replaceElements = [];

    replaceElements.forEach((elemName) => {
        let elems = document.getElementsByTagName(elemName);
        Array.from(elems).forEach((elem) => {
            const newElem = document.createElement('sd-' + elemName.toLowerCase());
            Array.from(elem.getAttributeNames()).forEach((attrName) => {
                newElem.setAttribute(attrName.toLowerCase(), elem.getAttribute(attrName))
            });
            newElem.innerHTML = elem.innerHTML;
            elem.parentNode.replaceChild(newElem, elem);

        });
    });
}