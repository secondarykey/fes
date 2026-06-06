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

function editTextArea(textArea) {
  var h = window.innerHeight;
  var th = h - 250;
  //大きさを決定
  textArea.style.height = th + "px";
  var rect = textArea.getBoundingClientRect();

  var main = document.querySelector("main");

  var pos = rect.y - 150;

  main.scrollTo(0,pos);
}

function confirmFes(msg,func) {
  if ( confirm(msg) ) {
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
