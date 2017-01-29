// ==UserScript==
// @name        Baldy Beaver
// @namespace   roc.baldy.beaver
// @description SPLoP API, aka splopi but better known as Baldy Beaver (Name was the collective effort of Sux0r and Whishy)
// @include     https://ruinsofchaos.com/*
// @exclude     https://ruinsofchaos.com/index.php*
// @exclude     https://ruinsofchaos.com/register.php*
// @exclude     https://ruinsofchaos.com/forgotpass.php*
// @version     1
// @grant       unsafeWindow
// @grant 		  GM_xmlhttpRequest
// @grant 		  GM_setValue
// @grant 		  GM_getValue
// @grant 		  GM_deleteValue
// @grant 		  GM_openInTab
// @grant 		  GM_addStyle
// ==/UserScript==

(function () {
  var $ = unsafeWindow.jQuery;
  
  var isChrome = /Chrome/.test(navigator.userAgent) && /Google Inc/.test(navigator.vendor);
  if (isChrome) {
      this.GM_getValue=function (key,def) {
          return localStorage[key] || def;
      };
      this.GM_setValue=function (key,value) {
          return localStorage[key]=value;
      };
      this.GM_deleteValue=function (key) {
          return delete localStorage[key];
      };
  }

  if(DetectRunningInstance() == true)
  {
      alert("You are running multiple instances of the BB script!\nManually remove the duplicates from Greasemonkey menu.");
  }
  
  if(document.body.innerHTML.indexOf('href="logout.php"') < 0 && document.body.innerHTML.indexOf('Rank') < 0 && document.body.innerHTML.indexOf('Turns') < 0){
    //User not logged in. No need to run the script.
    return;
  }else{
    addMenuPages();
  }
  
  var bbScriptServer = "http://127.0.0.1:8080";
  var scriptName = "Baldy Beaver";
  var url = document.location.toString();
  var BB_version = 1;
  var BB_server = "NA";
  
  var BB_username = GM_getValue("BB_username", "");
  var BB_password = GM_getValue("BB_password", "");
  var BB_statid = GM_getValue("BB_statid", "");
  var loggedIn = (GM_getValue("BB_SESSION_TYPE" , "") === "") ? "NODATA" : GM_getValue("BB_SESSION_TYPE" , "") ;
  
  
  if(BB_statid === ""){
    var userIdTok = document.getElementById("s_username").href.match(/stats\.php\?id=([0-9]*)/);
    BB_statid = userIdTok[1]
    GM_setValue("BB_statid" , BB_statid);
  }
  
  if(BB_username === ""){
    BB_username = document.getElementById("s_username").innerHTML;
    BB_username = BB_username.trim();
    GM_setValue("BB_username" , BB_username);
  }
  
  //Check if user is logged in before loading the scripts
  loadBBPage()
  
  function addMenuPages(){
    var bbMenu = $("<a class=\"bbMenu\" alt=\"Baldy Beaver\" href=\"base.php?bbpage=profile\"><span>VERY UGLY BUTTON</span></a>");
    var intelMenu = $("#menubar .menu7");
    intelMenu.after(bbMenu);
    bbMenu.css("display" , "block");
    bbMenu.css("height" , "20px");
    bbMenu.css("background-color" , "#432D12");
    bbMenu.css("padding" , "4px 0 0 0");
    bbMenu.css("color" , "#C3BC9F");
    bbMenu.css("font-weight" , "bold");
  }
  
  function loadBBPage(){
    if(url.indexOf("/base.php?bbpage=profile") > 0)
    {
      attemptLogin(bbBasePage);
    }
  }
  
  function bbBasePage(){
    var lolcontent = document.getElementById("lolcontent");
    lolcontent.innerHTML = "GETTING DATA, please hold on. Baldy script is balding stuff...";
    if(loggedIn === "NODATA"){
      var pageBory = "<div class=\"th topcap\">"+scriptName+"</div>";
      pageBory += "<table id=\"profileInfo\" class=\"sep\" width=\"100%\" cellspacing=\"0\" border=\"0\">";
      pageBory += "<tr class=\"odd\"><td colspan=\"3\">Please log in to continue. If you are not register, just chose a new password, and enter a random word for the challenge. Then message the challenge word in game to either Sux0r or Jellybean to be approved</td></tr>";
      pageBory += "<tr class=\"odd\"><td colspan=\"3\">You do not need to enter the challenge if you are already registered</td></tr>";
      pageBory += "<tr class=\"even\"><td><input type=\"text\" placeholder=\"Password\" id=\"inputPassword\" value=\""+BB_password+"\" /></td><td><input type=\"text\" placeholder=\"Challenge\" id=\"inputChallenge\"></td><td><button id=\"doLoginButton\">Login / Register</button></td></tr>";
      pageBory += "</table>";
      lolcontent.innerHTML = pageBory;
      var btn = document.getElementById("doLoginButton");
      btn.addEventListener('click' , doLogin );
    }else if (loggedIn === "LOGGEDIN"){
      getProfilePage();
    }
  }
  
  function getProfilePage(){
    GM_xmlhttpRequest({
      method: "POST",
      headers: { 'Content-type' : 'application/x-www-form-urlencoded' },
      data: encodeURI("password=" + BB_password + "&external_id=" + BB_statid),
      url: bbScriptServer + "/roc/getprofile",
      onload: function(r)
      {
        if(r.status == 200)
        {
          console.log(r.responseText);
        }
      }
    });
  }
  
  function doLogin(){
    var inputPassword = document.getElementById("inputPassword").value;
    var inputChallenge = document.getElementById("inputChallenge").value;
    
    GM_xmlhttpRequest({
      method: "POST",
      headers: { 'Content-type' : 'application/x-www-form-urlencoded' },
      data: encodeURI("username="+BB_username+"&password=" + inputPassword + "&external_id=" + BB_statid + "&challenge=" + inputChallenge),
      url: bbScriptServer + "/roc/sign",
      onload: function(r)
      {
        if(r.status == 200)
        {
          GM_setValue("BB_SESSION_TYPE" , r.responseText);
          if(r.responseText === "LOGGEDIN"){
            GM_setValue("BB_password" , inputPassword);
            BB_password = GM_getValue("BB_password", "");
			getProfilePage();
          } 
        }
      }
    });
    
  }
  
  function attemptLogin(callBack){
      //console.log("Attempting to log in");
      GM_xmlhttpRequest(
        {
          method: "POST",
          headers: { 'Content-type' : 'application/x-www-form-urlencoded' },
          data: encodeURI("username="+BB_username+"&password=" + BB_password + "&external_id=" + BB_statid + "&challenge" ),
          url: bbScriptServer + "/roc/sign",
          onload: function(r)
          {
            if(r.status == 200)
            {
				loggedIn = r.responseText ;
				callBack();
            }
          }
        });
  }
  
  function DetectRunningInstance()
  {
      if(document.getElementById('InstanceBB'))
      {
          return true;
      }

      var instanceDiv = document.createElement('div');
      instanceDiv.style.display = 'none';
      instanceDiv.setAttribute('id', "InstanceBB");
      document.body.appendChild(instanceDiv);

      return false;
  }
  
})()
