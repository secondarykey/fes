var xmlhttp;
if (window.XMLHttpRequest) { // code for IE7+, Firefox, Chrome, Opera, Safari
    xmlhttp=new XMLHttpRequest();
} else { // code for IE6, IE5
    xmlhttp=new ActiveXObject("Microsoft.XMLHTTP");
}

function EncodeHTMLForm(data){
    var params = [];
    for(var name in data){
        var value = data[name];
        var param = encodeURIComponent(name).replace(/%20/g, '+')
            + '=' + encodeURIComponent(value).replace(/%20/g, '+');
        params.push(param);
    }
    return params.join('&');
}

var singletonTextarea = true;
function editTextArea(textArea) {
    var dialog = document.querySelector('#textArea');
    if ( !dialog.showModal ) {
        dialogPolyfill.registerDialog(dialog);
    }
    var area = dialog.querySelector('#editTxt');
    area.value = textArea.value;

    if ( singletonTextarea ) {
        var close = dialog.querySelector('.close');
        var agree = dialog.querySelector('.agree');
        close.addEventListener('click', function() {
            dialog.close();
        });
        agree.addEventListener('click', function() {
            textArea.value = area.value;
            dialog.close();
        });
        singletonTextarea = false;
    }
    dialog.showModal();
}

function confirmFes(func) {
    if ( confirm("Realy?") ) {
          func();
    }
}

function alertFes(txt) {
    alert(txt);
}


(function() {
    var a, acc, i, len;
    acc = document.getElementsByClassName('accordion');
    for (i = 0, len = acc.length; i < len; i++) {
        a = acc[i];
        a.onclick = function() {
            this.classList.toggle('active');
            return this.nextElementSibling.classList.toggle('show');
        };
    }
}).call(this);