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
// @grant 		  GM_log
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
      var pageBody = "<div class=\"th topcap\">"+scriptName+"</div>";
      pageBody += "<table id=\"profileInfo\" class=\"sep\" width=\"100%\" cellspacing=\"0\" border=\"0\">";
      pageBody += "<tr class=\"odd\"><td colspan=\"3\">Please log in to continue. If you are not register, just chose a new password, and enter a random word for the challenge. Then message the challenge word in game to either Sux0r or Jellybean to be approved</td></tr>";
      pageBody += "<tr class=\"odd\"><td colspan=\"3\">You do not need to enter the challenge if you are already registered</td></tr>";
      pageBody += "<tr class=\"even\"><td><input type=\"text\" placeholder=\"Password\" id=\"inputPassword\" value=\""+BB_password+"\" /></td><td><input type=\"text\" placeholder=\"Challenge\" id=\"inputChallenge\"></td><td><button id=\"doLoginButton\">Login / Register</button></td></tr>";
      pageBody += "</table>";
      lolcontent.innerHTML = pageBody;
      var btn = document.getElementById("doLoginButton");
      btn.addEventListener('click' , doLogin );
    }else if(loggedIn === "WRONGPASS"){
	  var pageBody = "<div class=\"th topcap\">"+scriptName+"</div>";
      pageBody += "<table id=\"profileInfo\" class=\"sep\" width=\"100%\" cellspacing=\"0\" border=\"0\">";
      pageBody += "<tr class=\"odd\"><td colspan=\"3\">You either forgot your password or have fat fingers, you decide. Try again.</td></tr>";
      pageBody += "<tr class=\"even\"><td><input type=\"text\" placeholder=\"Password\" id=\"inputPassword\" value=\""+BB_password+"\" /></td><td><input type=\"text\" placeholder=\"Challenge\" id=\"inputChallenge\"></td><td><button id=\"doLoginButton\">Login / Register</button></td></tr>";
      pageBody += "</table>";
      lolcontent.innerHTML = pageBody;
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
			var rspObj = JSON.parse(r.responseText);
			loadUserMessage(rspObj);
			if(rspObj.TypeID === 4){
				loadNonAdminUsers(rspObj.UserEntries);
			}
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
      //GM_log("Attempting to log in");
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
  
  function changeStatus(id , newStatus){
	console.log(id , newStatus);
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
  
  function loadUserMessage(obj){
	//Loads the message from the server about the user type
	var lolcontent = document.getElementById("lolcontent");
	var loginStatus = (obj.Logged === true) ? "<span style=\"color : green\">Logged in</span>" : "<span style=\"color : red\">Not Logged</span><span style=\"font-size : 8px;font-weight:bold\">(Please contact bold_ally)</span>" ;
	var pageBody = "<div class=\"th topcap\">"+scriptName+"</div>";
	pageBody += "<table id=\"profileInfo\" class=\"sep\" width=\"100%\" cellspacing=\"0\" border=\"0\">";
	pageBody += "<tr class=\"odd\"><td width=\"50%\">"+loginStatus+"</td><td width=\"50%\">User status : "+obj.Type+"</td></tr>";
	pageBody += "</table>";
	lolcontent.innerHTML = pageBody;
  }
  
  function loadNonAdminUsers(userArray){
	var lolcontent = document.getElementById("lolcontent");
	var approvedTable = document.createElement("table");
	approvedTable.setAttribute("width" , "100%");
	approvedTable.setAttribute("cellspacing" , "0");
	approvedTable.setAttribute("border" , "0");
	var approvedTr = document.createElement("tr");
	var approvedTd = document.createElement("td");
	approvedTd.innerHTML="APPROVED LIST - Use button to remove approval";
	approvedTd.setAttribute("colspan" , "4");
	approvedTd.setAttribute("class" , "th topcap");
	approvedTr.appendChild(approvedTd);
	approvedTable.appendChild(approvedTr);
	
	var pendingTable = document.createElement("table");
	pendingTable.setAttribute("width" , "100%");
	pendingTable.setAttribute("cellspacing" , "0");
	pendingTable.setAttribute("border" , "0");
	var pendingTr = document.createElement("tr");
	var pendingTd = document.createElement("td");
	pendingTd.innerHTML="PENDING LIST - Use button to approve - Don't forget to check challenge first";
	pendingTd.setAttribute("colspan" , "4");
	pendingTd.setAttribute("class" , "th topcap");
	pendingTr.appendChild(pendingTd);
	pendingTable.appendChild(pendingTr);

	
	var approvedCounter = 0;
	var pendingCounter = 0;
	
	for(var index = 0 ; index < userArray.length ;  index++){
		userObj = userArray[index];
		var tr = document.createElement("tr");
		var span = document.createElement("span");
		var tdID = document.createElement("td");
		tdID.innerHTML = userObj.ExternalID;
		tdID.setAttribute("width" , "25%");
		var tdName = document.createElement("td");
		tdName.innerHTML = userObj.Username;
		tdName.setAttribute("width" , "25%");
		var tdChallenge = document.createElement("td");
		tdChallenge.innerHTML = userObj.Challenge;
		tdChallenge.setAttribute("width" , "25%");
		var tdChange = document.createElement("td");
		tdChange.appendChild(span);
		tdChange.setAttribute("width" , "25%");
		
		tr.appendChild(tdID);
		tr.appendChild(tdName);
		tr.appendChild(tdChallenge);
		tr.appendChild(tdChange);
		
		
		if(userObj.TypeID === 2){
			span.innerHTML = "[&#10004; - Approve]";
			span.setAttribute("style" , "color:#159615 ; font-weight : bold; font-size : 14px ; cursor : pointer;" );
			span.addEventListener("click" , function(id , newStatus){changeStatus(id , newStatus)}.bind(span , userObj.ExternalID , 3));
			pendingTable.appendChild(tr);
			if(pendingCounter % 2 == 0){
				tr.setAttribute("class" , "even");
			}else{
				tr.setAttribute("class" , "odd");
			}
			pendingCounter++;

		}else if(userObj.TypeID === 3){
			span.innerHTML = "[&#10008; - Remove Access - Contact bold_ally to remove from DB]";
			span.setAttribute("style" , "color:#b50101 ; font-weight : bold; font-size : 14px; cursor : pointer;" );
			span.addEventListener("click" , function(id , newStatus){changeStatus(id , newStatus)}.bind(span , userObj.ExternalID , 2));
			approvedTable.appendChild(tr);
			if(approvedCounter % 2 == 0){
				tr.setAttribute("class" , "even");
			}else{
				tr.setAttribute("class" , "odd");
			}
			approvedCounter++;
		}
	}
	lolcontent.appendChild(approvedTable);
	lolcontent.appendChild(pendingTable);
  }
  
  
})()
