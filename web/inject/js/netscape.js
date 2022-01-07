function initNetscape() {

    let fontElements = document.getElementsByTagName("font");
    fontElements.forEach(function (font) {
        const shadow = font.attachShadow({mode: 'open'});
        let fontFace = font.getAttribute('face');
        let fontFamily = fontFace;
        switch (fontFace) {
            case 'Courier':
                fontFamily = 'Unifont';
                break
        }

        // CSS anlegen und ins Schatten-Dom einhängen
        // :host selektiert das Custom Element
        const style = document.createElement('style');
        style.textContent = `
			:host {
			    font-family: '${fontFamily}';
			}	
		`;
        shadow.appendChild(style);
    });
}

// Custom-Element my-element anlegen
class xFont extends HTMLElement {

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
        const shadow = this.attachShadow({mode: 'open'});

        /*
        // Element für Inhalt anlegen und ins Schatten-Dom einhängen
        const content = document.createElement('div');
        content.className = "content";
        shadow.appendChild(content);
         */


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

        if( name == 'face' ) {
            let fontFamily = newValue;
            switch (newValue) {
                case 'Courier':
                    fontFamily = 'Unifont';
                    break
            }

            // CSS anlegen und ins Schatten-Dom einhängen
            // :host selektiert das Custom Element
            const style = document.createElement('style');
            style.textContent = `
			:host {
			    font-family: '${fontFamily}';
			}	
		`;
            shadow.appendChild(style);
        }
    }

}
//customElements.define('x-font', xFont);