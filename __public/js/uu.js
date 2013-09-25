var hljs_languages;

var cur_lang;
var password;

function updateHash()
{
  var hashS ="";
  if(password)
  {
    hashS = hashS + "clef=" + password;
  }
  if(cur_lang)
  {
    hashS = hashS + "&lang=" + cur_lang;
  }
  document.location.hash = hashS;
}

function changeToLink(lang) {
  return '<a class="clickable" onclick="changeTo(\''+lang +'\');">'+lang+'</a>';
}

function generatePassword(length) {
  var pass=Math.abs(sjcl.random.randomWords(1)[0]);
  var p = "";
  while(p.length < length) {
    p = p + String.fromCharCode(97 + (pass % 26));
    pass = pass / 26;
    if(pass < 1) {
      pass = Math.abs(sjcl.random.randomWords(1)[0]);
    }
  }
  return p;
}
function parseHash(hsh) {
  p = hsh.split("&");
  parms = {};
  for(i=0; i < p.length; i++) {
    kv = p[i].split("=")
    parms[kv[0]] = kv[1];
  }
  return parms;
}

function displayError() {
  setPasteto("Invalid key in url. Please make sure # parameter contains clef value");
}

var parms = parseHash(window.location.hash.substr(1));
var password;

function wakeUp() {

  if(parms['clef']) {
    password = parms['clef'];
    var data = $('#encrypted').text();
    var atts = $('#attachments').text();

    try {
      var ct = sjcl.decrypt(password, data);
      setPasteto(ct);
    } catch(e) {
      if(e instanceof sjcl.exception.corrupt) {
        displayError();
      }
    } 
    setAttachmentsTo(atts);
  }else {
    displayError();
  }
  $('#reveal-but').click(revealRaw);
  $('#wrap-but').click(wrapidou);
  $('#code').show();
  $('#spare').hide();
  $('#spare').height(80*$(window).height()/100);


}
function RNGisReady() {
  $('#gl-but').removeAttr("disabled");
  $('#gl-text').text("");
  if(!password)
    password = generatePassword(16);

}

if(encrypted) {
  $(document).ready(function() {wakeUp()});
}


$(document).ready(function() {
  sjcl.random.startCollectors();
  if(sjcl.random.isReady()) {
    RNGisReady();
  } else {
    sjcl.random.addEventListener("seeded", function(){
      RNGisReady();
    });
  }

  hljs_languages = Object.keys(hljs.LANGUAGES);
  $('#placeholder').replaceWith(function() {
    var o ='';
    for(i=0; i < hljs_languages.length; i++) {
      if(hljs_languages[i] == cur_lang)
      {
        continue;
      }
      o = o + changeToLink(hljs_languages[i]) +" ";
    }
    return o
  });

  $('#never_expire').change( function(o) {
    var state = $(this).is(':checked');
    if(state)
    {
      $('#expiry_delay').attr('disabled', 'disabled');
    } else {
      $('#expiry_delay').removeAttr('disabled');
    }
  });

  $("#gluu_form").submit(function () {
    var data = $("#clear_code").val();
    if(!sjcl.random.isReady())
    {
      alert("RNG not ready, move you mouse a bit and try again.");
      return false;
    }

    var encrypted = sjcl.encrypt(password, data);

    var sent_data = { content: encrypted, attachments: $("#attachments").text() };

    if($('#never_expire').is(':checked')) {
      sent_data['never_expire'] = true;
    } else {
      sent_data['expiry_delay'] = $('#expiry_delay').val();
    }

    $.ajax({
      type: "POST",
      url: "/paste",
      data: sent_data
    }).done(function( dest_url ) {
      document.location.href = dest_url + "#clef="+password;
    }).fail(function(jqXHR, textStatus) {
      console.log( "Request failed: " + textStatus );
    });

    return false;

  });

  var elt = $("#code")[0];
  if(elt) {
    hljs.highlightBlock(elt);

  }

  if(parms['lang']) {
    changeTo(parms['lang']);
  }

  Dropzone.options.uploader = {
  paramName: "file", // The name that will be used to transfer the file
  maxFilesize: 10, // MB
  success: function(file, text) {
    var attachment_name = "Attachment:/a/"+text+"\n";
    $("#attachments").append(attachment_name);
    return file.previewElement.classList.add("dz-success");
  },
  init: function() {
    this.on("sending", function(file, xhr, formdata) {
      var expiry;
      if($('#never_expire').is(':checked')) {
        formdata.append('never_expire',true);
      } else {
        formdata.append('expiry_delay',$('#expiry_delay').val());
      }
    });
  }
};

 });

function loadCssAtIdentifier(cssName, identifier)
{
  var link = $("<link>");

  link.attr({ id: identifier,
          type: 'text/css',
          rel: 'stylesheet',
          href: cssName +'.css'
  });
  $("#" + identifier).remove();
  $("head").append( link ); 
}

function setPasteto(content, lang) {
  var langClass = "";
  if(lang != undefined) {
    langClass = "class='"+lang+"'";
  }
  $('#code').replaceWith(function() {
    var c = $("<code id='code' "+langClass+"></code>");
    c.text(content);
    return c;
  });
  hljs.highlightBlock($('#code')[0]);

  if(lang == undefined) {
    var detected = $($("#code")[0]).attr('class');
    cur_lang = detected;
    $('#hljs_lang').text(detected);
  }
  number();
  $('#spare').text(content);

}

function setAttachmentsTo(content) {
  var attn = content.split(/\n/);
  for(var i=0; i < attn.length; i++){
    var currentA = attn[i];
    if(/Attachment:/.test(currentA))
    {
      if(/(.jpeg$|.jpg$|.png|.gif)$/.test(currentA)) {
        $('#companions').append($('<img />', { 'src': currentA.substr(11), 'class' : 'attachment' }));
      } else {
        $('#companions').append($('<a />', { 'text':currentA, 'href': currentA.substr(11), 'class' : 'attachment' }));
      }
    } else {
    }
  }
}

function changeTo(lang) {
  setPasteto($('#spare').text(), lang);
  $('#hljs_lang').text(lang);
  cur_lang = lang;
  updateHash();
}
function revealRaw() {
  $('#spare').toggle();
  $('#code').toggle();
  $('#spare').select();
}

function wrapidou() {
  $('pre').toggleClass("wrap");
}

function number() {
$('pre>code').each(function() {
    var current = $(this),
        lineStart = parseInt(current.data('line-start')),
        lineFocus = parseInt(current.data('line-focus')),
        items = current.html().split("\n"),
        total = items.length,
        result = '<ol ' + (!isNaN(lineStart) ? 'start="' + lineStart + '"' : '') + '>';

    for (var i = 0; i < total; i++) {
        if (i === (lineFocus - lineStart)) {
            result += '<li class="hightlight">';
        }
        else {
            result += '<li>';
        }

        result += items[i] + '</li>';
    }

    result += '</ol>';

    var items = current.empty().append(result);
});

}