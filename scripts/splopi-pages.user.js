// ==UserScript==
// @name        Baldy Beaver
// @namespace   roc.baldy.beaver
// @description SPLoP API, aka splopi but better known as Baldy Beaver (Name was the collective effort of Sux0r and Whishy)
// @include     https://ruinsofchaos.com/*
// @exclude     https://ruinsofchaos.com/index.php*
// @exclude     https://ruinsofchaos.com/register.php*
// @exclude     https://ruinsofchaos.com/forgotpass.php*
// @version     1.03
// @grant 		  GM_xmlhttpRequest
// @grant 		  GM_setValue
// @grant 		  GM_getValue
// @grant 		  GM_deleteValue
// @grant 		  GM_openInTab
// @grant 		  GM_addStyle
// @grant 		  GM_log
// @require 	https://code.jquery.com/jquery-2.2.4.min.js
// ==/UserScript==

(function () {
	var isChrome = /Chrome/.test(navigator.userAgent) && /Google Inc/.test(navigator.vendor);
	if (isChrome) {
		this.GM_getValue = function (key, def) {
			return localStorage[key] || def;
		};
		this.GM_setValue = function (key, value) {
			return localStorage[key] = value;
		};
		this.GM_deleteValue = function (key) {
			return delete localStorage[key];
		};
	}

	if (DetectRunningInstance() == true) {
		alert("You are running multiple instances of the BB script!\nManually remove the duplicates from Greasemonkey menu.");
	}

	if (document.body.innerHTML.indexOf('href="logout.php"') < 0 && document.body.innerHTML.indexOf('Rank') < 0 && document.body.innerHTML.indexOf('Turns') < 0) {
		//User not logged in. No need to run the script.
		return;
	} else {
		addMenuPages();
	}

	var bbScriptServer = "http://52.10.254.235:8080";
	//var bbScriptServer = "http://127.0.0.1:8080";
	
	var scriptName = "Baldy Beaver";
	var url = document.location.toString();
	var BB_version = 1;
	var BB_server = "NA";

	var BB_username = GM_getValue("BB_username", "");
	var BB_password = GM_getValue("BB_password", "");
	var BB_statid = GM_getValue("BB_statid", "");
	var loggedIn = (GM_getValue("BB_SESSION_TYPE", "") === "") ? "NODATA" : GM_getValue("BB_SESSION_TYPE", "");

	if (BB_statid === "") {
		var userIdTok = document.getElementById("s_username").href.match(/stats\.php\?id=([0-9]*)/);
		BB_statid = userIdTok[1]
			GM_setValue("BB_statid", BB_statid);
	}

	if (BB_username === "") {
		BB_username = document.getElementById("s_username").innerHTML;
		BB_username = BB_username.trim();
		GM_setValue("BB_username", BB_username);
	}

	//Check if user is logged in before loading the scripts
	loadBBPage()

	function addMenuPages() {
		var bbMenu = $("<a class=\"bbMenu\" alt=\"Baldy Beaver\" href=\"base.php?bbpage=profile\"><span>VERY UGLY BUTTON</span></a>");
		var intelMenu = $("#menubar .menu7");
		intelMenu.after(bbMenu);
		bbMenu.css("display", "block");
		bbMenu.css("height", "20px");
		bbMenu.css("background-color", "#432D12");
		bbMenu.css("padding", "4px 0 0 0");
		bbMenu.css("color", "#C3BC9F");
		bbMenu.css("font-weight", "bold");
		
		var sabMenu = $("<a class=\"sabMenu\" alt=\"Baldy Beaver\" href=\"base.php?bbpage=sablist\"><span>SAB LIST!</span></a>");
		var intelMenu = $("#menubar .menu7");
		intelMenu.after(sabMenu);
		sabMenu.css("display", "block");
		sabMenu.css("height", "22px");
		sabMenu.css("background-color", "#333333");
		sabMenu.css("padding", "4px 0 0 0");
		sabMenu.css("color", "#FF0000");
		sabMenu.css("font-weight", "bold");
		sabMenu.css("font-size", "15px");
	}

	function loadBBPage() {
		if (url.indexOf("/base.php?bbpage=profile") > 0) {
			attemptLogin(bbBasePage);
		} else if (url.indexOf("/base.php?bbpage=sablist") > 0) {
			getSabList();
		} else if (url.indexOf("/inteldetail.php") > 0) {
			inteldetail();
		} else if (url.indexOf("/stats.php") > 0) {
			stats();
		}
	}

	function getSabList(){
		var lolcontent = document.getElementById("lolcontent");
		lolcontent.innerHTML = "GETTING WAR DATA!!!!";
		
		GM_xmlhttpRequest({
			method: "POST",
			headers: {
				'Content-type': 'application/x-www-form-urlencoded'
			},
			data: encodeURI("external_id=" + BB_statid),
			url: bbScriptServer + "/roc/getsablist",
			onload: function (r) {
				if (r.status == 200) {
					var playerArray = JSON.parse(r.responseText);
					loadSabList(playerArray);
				}
			}
		});
		
	}
	
	function bbBasePage() {
		var lolcontent = document.getElementById("lolcontent");
		lolcontent.innerHTML = "GETTING DATA, please hold on. Baldy script is balding stuff...";
		if (loggedIn === "NODATA") {
			var pageBody = "<div class=\"th topcap\">" + scriptName + "</div>";
			pageBody += "<table id=\"profileInfo\" class=\"sep\" width=\"100%\" cellspacing=\"0\" border=\"0\">";
			pageBody += "<tr class=\"odd\"><td colspan=\"3\">Please log in to continue. If you are not register, just chose a new password, and enter a random word for the challenge. Then message the challenge word in game to either Sux0r or Jellybean to be approved</td></tr>";
			pageBody += "<tr class=\"odd\"><td colspan=\"3\">You do not need to enter the challenge if you are already registered</td></tr>";
			pageBody += "<tr class=\"even\"><td><input type=\"text\" placeholder=\"Password\" id=\"inputPassword\" value=\"" + BB_password + "\" /></td><td><input type=\"text\" placeholder=\"Challenge\" id=\"inputChallenge\"></td><td><button id=\"doLoginButton\">Login / Register</button></td></tr>";
			pageBody += "</table>";
			lolcontent.innerHTML = pageBody;
			var btn = document.getElementById("doLoginButton");
			btn.addEventListener('click', doLogin);
		} else if (loggedIn === "WRONGPASS") {
			var pageBody = "<div class=\"th topcap\">" + scriptName + "</div>";
			pageBody += "<table id=\"profileInfo\" class=\"sep\" width=\"100%\" cellspacing=\"0\" border=\"0\">";
			pageBody += "<tr class=\"odd\"><td colspan=\"3\">You either forgot your password or have fat fingers, you decide. Try again.</td></tr>";
			pageBody += "<tr class=\"even\"><td><input type=\"text\" placeholder=\"Password\" id=\"inputPassword\" value=\"" + BB_password + "\" /></td><td><input type=\"text\" placeholder=\"Challenge\" id=\"inputChallenge\"></td><td><button id=\"doLoginButton\">Login / Register</button></td></tr>";
			pageBody += "</table>";
			lolcontent.innerHTML = pageBody;
			var btn = document.getElementById("doLoginButton");
			btn.addEventListener('click', doLogin);
		} else if (loggedIn === "LOGGEDIN") {
			getProfilePage();
		}
	}

	function getProfilePage() {
		GM_xmlhttpRequest({
			method: "POST",
			headers: {
				'Content-type': 'application/x-www-form-urlencoded'
			},
			data: encodeURI("password=" + BB_password + "&external_id=" + BB_statid),
			url: bbScriptServer + "/roc/getprofile",
			onload: function (r) {
				if (r.status == 200) {
					var rspObj = JSON.parse(r.responseText);
					loadUserMessage(rspObj);
					if (rspObj.TypeID === 4) {
						loadNonAdminUsers(rspObj.UserEntries);
					}
				}
			}
		});
	}

	function doLogin() {
		var inputPassword = document.getElementById("inputPassword").value;
		var inputChallenge = document.getElementById("inputChallenge").value;

		GM_xmlhttpRequest({
			method: "POST",
			headers: {
				'Content-type': 'application/x-www-form-urlencoded'
			},
			data: encodeURI("username=" + BB_username + "&password=" + inputPassword + "&external_id=" + BB_statid + "&challenge=" + inputChallenge),
			url: bbScriptServer + "/roc/sign",
			onload: function (r) {
				if (r.status == 200) {
					GM_setValue("BB_SESSION_TYPE", r.responseText);
					if (r.responseText === "LOGGEDIN") {
						GM_setValue("BB_password", inputPassword);
						BB_password = GM_getValue("BB_password", "");
						getProfilePage();
					}
				}
			}
		});

	}

	function attemptLogin(callBack) {
		//GM_log("Attempting to log in");
		GM_xmlhttpRequest({
			method: "POST",
			headers: {
				'Content-type': 'application/x-www-form-urlencoded'
			},
			data: encodeURI("username=" + BB_username + "&password=" + BB_password + "&external_id=" + BB_statid + "&challenge"),
			url: bbScriptServer + "/roc/sign",
			onload: function (r) {
				if (r.status == 200) {
					loggedIn = r.responseText;
					callBack();
				}
			}
		});
	}

	function changeStatus(span, id, newStatus) {
		GM_xmlhttpRequest({
			method: "POST",
			headers: {
				'Content-type': 'application/x-www-form-urlencoded'
			},
			data: encodeURI("external_id=" + BB_statid + "&user_id=" + id + "&type_id=" + newStatus),
			url: bbScriptServer + "/roc/changetypeid",
			onload: function (r) {
				if (r.status == 200) {
					console.log(r.status, r.responseText);
					span.innerHTML = "[STATUS CHANGED - RELOAD PAGE]";
					span.setAttribute("style", "color:#FFFFFF ; font-weight : bold; font-size : 14px ; cursor : default;");
				} else {
					document.getElementById("lolcontent").innerHTML = "STOP TRYING TO BREAK MY SCRIPT!!!, or might need to login again. I can't really tell";
				}
			}
		});
	}

	function DetectRunningInstance() {
		if (document.getElementById('InstanceBB')) {
			return true;
		}

		var instanceDiv = document.createElement('div');
		instanceDiv.style.display = 'none';
		instanceDiv.setAttribute('id', "InstanceBB");
		document.body.appendChild(instanceDiv);

		return false;
	}

	function loadUserMessage(obj) {
		//Loads the message from the server about the user type
		var lolcontent = document.getElementById("lolcontent");
		var loginStatus = (obj.Logged === true) ? "<span style=\"color : green\">Logged in</span>" : "<span style=\"color : red\">Not Logged</span><span style=\"font-size : 8px;font-weight:bold\">(Please contact bold_ally)</span>";
		var pageBody = "<div class=\"th topcap\">" + scriptName + "</div>";
		pageBody += "<table id=\"profileInfo\" class=\"sep\" width=\"100%\" cellspacing=\"0\" border=\"0\">";
		pageBody += "<tr class=\"odd\"><td width=\"50%\">" + loginStatus + "</td><td width=\"50%\">User status : " + obj.Type + "</td></tr>";
		pageBody += "</table>";
		lolcontent.innerHTML = pageBody;
	}

	function loadNonAdminUsers(userArray) {
		var lolcontent = document.getElementById("lolcontent");
		var approvedTable = document.createElement("table");
		approvedTable.setAttribute("width", "100%");
		approvedTable.setAttribute("cellspacing", "0");
		approvedTable.setAttribute("border", "0");
		var approvedTr = document.createElement("tr");
		var approvedTd = document.createElement("td");
		approvedTd.innerHTML = "APPROVED LIST - Use button to remove approval";
		approvedTd.setAttribute("colspan", "4");
		approvedTd.setAttribute("class", "th topcap");
		approvedTr.appendChild(approvedTd);
		approvedTable.appendChild(approvedTr);

		var pendingTable = document.createElement("table");
		pendingTable.setAttribute("width", "100%");
		pendingTable.setAttribute("cellspacing", "0");
		pendingTable.setAttribute("border", "0");
		var pendingTr = document.createElement("tr");
		var pendingTd = document.createElement("td");
		pendingTd.innerHTML = "PENDING LIST - Use button to approve - Don't forget to check challenge first";
		pendingTd.setAttribute("colspan", "4");
		pendingTd.setAttribute("class", "th topcap");
		pendingTr.appendChild(pendingTd);
		pendingTable.appendChild(pendingTr);

		var approvedCounter = 0;
		var pendingCounter = 0;

		for (var index = 0; index < userArray.length; index++) {
			userObj = userArray[index];
			var tr = document.createElement("tr");
			var span = document.createElement("span");
			var tdID = document.createElement("td");
			tdID.innerHTML = userObj.ExternalID;
			tdID.setAttribute("width", "25%");
			var tdName = document.createElement("td");
			tdName.innerHTML = userObj.Username;
			tdName.setAttribute("width", "25%");
			var tdChallenge = document.createElement("td");
			tdChallenge.innerHTML = userObj.Challenge;
			tdChallenge.setAttribute("width", "25%");
			var tdChange = document.createElement("td");
			tdChange.appendChild(span);
			tdChange.setAttribute("width", "25%");

			tr.appendChild(tdID);
			tr.appendChild(tdName);
			tr.appendChild(tdChallenge);
			tr.appendChild(tdChange);

			if (userObj.TypeID === 2) {
				span.innerHTML = "[&#10004; - Approve]";
				span.setAttribute("style", "color:#159615 ; font-weight : bold; font-size : 14px ; cursor : pointer;");
				span.addEventListener("click", function (id, newStatus) {
					changeStatus(this, id, newStatus)
				}
					.bind(span, userObj.ExternalID, 3));
				pendingTable.appendChild(tr);
				if (pendingCounter % 2 == 0) {
					tr.setAttribute("class", "even");
				} else {
					tr.setAttribute("class", "odd");
				}
				pendingCounter++;

			} else if (userObj.TypeID === 3) {
				span.innerHTML = "[&#10008; - Remove Access - Contact bold_ally to remove from DB]";
				span.setAttribute("style", "color:#b50101 ; font-weight : bold; font-size : 14px; cursor : pointer;");
				span.addEventListener("click", function (id, newStatus) {
					changeStatus(this, id, newStatus)
				}
					.bind(span, userObj.ExternalID, 2));
				approvedTable.appendChild(tr);
				if (approvedCounter % 2 == 0) {
					tr.setAttribute("class", "even");
				} else {
					tr.setAttribute("class", "odd");
				}
				approvedCounter++;
			}
		}
		lolcontent.appendChild(approvedTable);
		lolcontent.appendChild(pendingTable);
	}

	//Sell off value part
	function inteldetail() {
		/*
		 * This grabs the first part of the page that inculdes information about the report
		 ********************************************************************************
		 * EX: Your spy sneaks into someone's base to attempt to gather intelligence.   *
		 * Your spy enters undetected. The spy gathers the following information:       *
		 ********************************************************************************
		 * Using this information we will determin if the intel report is a successful recon,
		 * or something else we don't care about
		 */
		//Local variables
		//This will store all the weapons for that user. Will also be used to store in a cookie for use outside of the intel page
		var weaponsArray = [
		];
		//The current estimated sell value without calculating the weapons that cannot be  indentified
		var cesov = 0;
		//This grabs the acutal object and not the text
		//This might be bad cause it depends on the class. Will figure out a way the description using a different method
		var descriptionObject = $('.td.c');
		//Grab the text from the object
		var descriptionText = descriptionObject.html();
		/**
		 * Sample out put of the descriptionText
		 * console.log(descriptionText);

		Your 10 spies sneak into <a href="stats.php?id=28380">someone</a>'s base to attempt to gather intelligence.<br>
		<br>
		Your spies enter undetected.
		<font size="3" color="RED">2</font> of your spies were a little slow and had to become an hero in order to evade capture and interrogation.
		The spies gather the following information:

		 * We will be using the use ID to reference the player
		 **/
		//Search for "gathers the following information" to determin if this is the info we need
		if (descriptionText.indexOf('gathers the following information') == 0) {
			//This is not a valid log, and we need to stop the script,
			//will comment out the log to keep things clean
			//console.log("This is not a valid intel log");
			//Return will cause the function to finish executing
			return;
		} //Use regular expression to match and extract the userID

		var userIDregex = /stats\.php\?id=([0-9]+)/;
		var match = descriptionText.match(userIDregex);
		if (match.length < 2) {
			//Either code has changed, or I did something wrong, end the scrip for now
			return;
		} //We store the user id, it will be later used to store the remaining values
		var userID = match[1];

		//Now we grab the weapons type
		//We start by grabbig all the tables from the page
		var tables = $('table.sep');
		//Iterate over all the retrieved table to find the one with the header title "Weapons"
		var weaponsObj = false;

		tables.each(function () {
			//make a jQuery object
			var tableObj = $(this);
			var tbodyObj = tableObj.children();
			//Check if the first element is a table body
			//I do not think this is needed. And that check can be bypassed
			if (tbodyObj[0].localName == 'tbody') {
				//We now fetch the first row from the table
				var trObj = tbodyObj.children().first();
				//Each row is formed of multiple cells, in the case of the header
				//It only has one cell, that contains the title
				//We fetch the first cell and check the title name
				var headerCell = trObj.children().first();
				//We check the value inside to figure out if this is the table we need
				//3 equals also matches the type
				if (headerCell.html() === 'Weapons') {
					//Once we found the Weapons table, we store the object and return the inner function
					//This return will not end the entire script, it will only end the function we are currently in
					weaponsObj = tableObj;
					return;
				}
			}

		});

		//close the script if we did not find the Weapons table
		if (weaponsObj === false) {
			return;
		} //Get rows from the table.
		//The structure is as follow, Table > tablebody > rows > cells. Also skip the first 2 elemnts since they are headers.
		//gt is zero based, gt(0) will skip the first, gt(1) will skip the first 2

		var weaponRows = weaponsObj.children().children(':gt(1)');
		//Get the first row for the title, we will use that later to output our sell off value
		var titleRow = weaponsObj.children().children(':nth-child(1)');
		//Get the titleCell
		titleCell = $(titleRow.children()[0]);
		//Looping over all the rows, each row is a description of a weapon held by user
		weaponRows.each(function () {
			//Create object from row
			var weaponRow = $(this);
			//Get the cells from the row, there are 4 cells, name, type, quantity and strength
			var weaponCells = weaponRow.children();
			//Create an object of the cell, then extract the value from each cell.
			var name = $(weaponCells[0]).html();
			//We need to strip the link from the name that links to a sab. Regular expressions are your friend
			//Sample name text without the link stripped
			//<a href="attack.php?sab=WhoCares&amp;weapon=Maul">Maul</a>
			//Keep in mind that the value can be "???" we need to account for that. I used ternary  operator here
			name = (name === '???') ? '???' : name.match(/>(.*)<\/a>/)[1];
			var type = $(weaponCells[1]).html();
			//Remove the comma from the quantity and transform the string to an integer
			var quantity = ($(weaponCells[2]).html() === '???') ? -1 : parseInt(replaceAll($(weaponCells[2]).html(), ',', ''));
			var curStrength = $(weaponCells[3]).html();
			//Splitting the curStrength to get the current damage and the max strength
			//The max strength will help us determin the sell value of the item even if we are lacking information
			//Check for ??? as well.
			//We add place holder values
			var damage = -1;
			var strength = -1;
			if (curStrength !== '???') {
				var strengthTok = curStrength.split('/');
				damage = strengthTok[0];
				strength = strengthTok[1];
			} //Will now create an object for this weapon, this object will be used to store it in a cookie for later use.

			var wep = {
				'name': name,
				'type': type,
				'quantity': quantity,
				'damage': damage,
				'strength': strength
			};
			//Add that object to the weapon array
			weaponsArray.push(wep);

		});
		var webSTR = JSON.stringify(weaponsArray);
		//Send the weapon array as an string to the server.
		//We also do not care about the result. The server will decide what to do with the data.
		//The results will show on the user page

		GM_xmlhttpRequest({
			method: "POST",
			headers: {
				'Content-type': 'application/x-www-form-urlencoded'
			},
			data: encodeURI("external_id=" + BB_statid + "&user_id=" + userID + "&weapons=" + webSTR),
			url: bbScriptServer + "/roc/storesov",
			onload: function (r) {
				//Do nothing
			}
		});

	}

	function stats() {
		var pageURL = window.location.href;
		var userIdTok = pageURL.match(/stats\.php\?id=([0-9]*)/);
		if (userIdTok.length === 2) {
			var userID = userIdTok[1];

			GM_xmlhttpRequest({
				method: "POST",
				headers: {
					'Content-type': 'application/x-www-form-urlencoded'
				},
				data: encodeURI("external_id=" + BB_statid + "&user_id=" + userID),
				url: bbScriptServer + "/roc/getplayerpage",
				onload: function (r) {
				
					var obj = JSON.parse(r.responseText);
					var tables = $(".sep.f");
					var goldTable = tables[1];
					var gtObj = $(goldTable);
					var pObj = obj.Player;
					if (pObj.Sov !== "") {
						var sov = pObj.Sov;
						var sovRow = $('<tr><td class="lg" style="width: 25%"><b>SOV:</b></td><td class="lg" style="width: 75%">' + sov.toLocaleString() + '</td></tr>');
						gtObj.children().append(sovRow);
					}
					if (pObj.Note !== "") {
						var note = pObj.Note
						var noteRow = $('<tr><td class="lg" style="width: 25%"><b>Note:</b></td><td class="lg" style="width: 75%">' + note + '</td></tr>');
						gtObj.children().append(noteRow);					
					}

					if(obj.Type === 4){
						var addNoteBtn = $("<button id=\"updateBTN\">Update</button>");
						addNoteBtn.on("click" , updateUserNote) ;
						var  tr = $('<tr></tr>');
						var td = $('<td class="lg" colspan="2"><input type="text" style="width:75%;" id="updatenode" ></td>');
						td.append(addNoteBtn);
						addNoteBtn.css("width" , "23%");
						tr.append(td);
						gtObj.children().append(tr);
					}
					
				}
			});

		} else {
			console.log("STOP TRYING TO BREAK MY SCRIPT");
		}
	}
	
	function updateUserNote(){
		var pageURL = window.location.href;
		var userIdTok = pageURL.match(/stats\.php\?id=([0-9]*)/);
		if (userIdTok.length === 2) {
			var userID = userIdTok[1];
			
			var tables = $(".sep.f");
			var userInfoTb = tables[0];
			var userInfoth = $(userInfoTb).find("th");
			var playerName = userInfoth[0].innerHTML;

			var note = $("#updatenode").val();
			
			GM_xmlhttpRequest({
				method: "POST",
				headers: {
					'Content-type': 'application/x-www-form-urlencoded'
				},
				data: encodeURI("external_id=" + BB_statid + "&user_id=" + userID + "&playername=" + playerName + "&note=" + note),
				url: bbScriptServer + "/roc/updateNote",
				onload: function (r) {
					if (r.status == 200) {
						$("#updateBTN").html("UPDATED!");
						$("#updateBTN").css("background-color" ,"#aaffff"); 
					}
				}
			});
			
				
		}
	}

	function replaceAll(target, search, replacement) {
		return target.split(search).join(replacement);
	};

	function loadSabList(players){
		var lolcontent = document.getElementById("lolcontent");
		var sabTable = document.createElement("table");
		sabTable.setAttribute("width", "100%");
		sabTable.setAttribute("cellspacing", "0");
		sabTable.setAttribute("border", "0");

		var headerTr = document.createElement("tr");
		var headerTd = document.createElement("td");
		headerTd.innerHTML = "APPROVED SAB LIST";
		headerTd.setAttribute("colspan", "4");
		headerTd.setAttribute("class", "th topcap");
		headerTr.appendChild(headerTd);
		sabTable.appendChild(headerTr);
		
		var counter = 0;

		for (var index = 0; index < players.length; index++) {
			userObj = players[index];
			
			var tr = document.createElement("tr");
			
			var tdName = document.createElement("td");
			tdName.innerHTML = '<a href="https://ruinsofchaos.com/stats.php?id='+userObj.ExternalID+'">' + userObj.Name + '</a>';
			tdName.setAttribute("width", "25%");
			
			var tdNote = document.createElement("td");
			tdNote.innerHTML = getSabLinks(userObj.Note , userObj.Name);
			tdNote.setAttribute("width", "75%");

			tr.appendChild(tdName);
			tr.appendChild(tdNote);
			
			if (counter % 2 == 0) {
				tr.setAttribute("class", "even");
			} else {
				tr.setAttribute("class", "odd");
			}
			sabTable.appendChild(tr);
			counter++;
			
		}
		lolcontent.innerHTML = "";
		lolcontent.appendChild(sabTable);
		
	}
	
	function getSabLinks(note , name){
		var weapons = ["Dagger" , "Maul" , "Blade" , "Excalibur" , "Sai" , "Shield" , "Mithril" , "Dragonskin" , "Cloak" , "Hook" , "Pickaxe" , "Horn" , "Guard Dog" , "Torch"];
		
		for (var index in weapons){
			note = note.replace(weapons[index] , '<a href="https://ruinsofchaos.com/attack.php?sab='+name+'&weapon='+weapons[index]+'">'+weapons[index]+'</a>' )
		}
		
		return note;
	}
	
})()
