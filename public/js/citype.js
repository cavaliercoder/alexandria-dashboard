var selectedAtt = null;

var inputTypeName = null;
var inputTypeDescription = null;
var ulAtts = null;
var liNewAtt = null;
var liAtt = null;
var inputAttName = null;
var inputAttDescription = null;
var selectAttType = null;

var submitForm = null;
var submitData = null;
var buttonSave = null;

var suspendUi = false;

function addAttribute() {
	//Find next available 'New Attribute' name
	var name = 'New Attribute';
	for (i = 1; i < 1024; i++) {
		var test = name;
		if (i > 1) test += ' ' + i;
		var matched = false;
		for (a = 0; a < citype.attributes.length; a++) {
			if (citype.attributes[a].name == test) {
				matched = true;
			}
		}

		if (!matched) {
			name = test;
			break;
		}
	}

	// Add a basic attribute
	var att = {
		"name" : name,
		"description": "",
		"type": "string"
	};
	citype.attributes.push(att);

	// Add a list item
	liAtt = buildAttributeListItem(att, null);
	ulAtts.append(liAtt);

	setAttribute(att);

	// Send user focus to the attribute name field
	inputAttName.focus();
	inputAttName.select();
}

function buildAttributeListItem(att, li) {
	if (! li) {
		li = $('<li class="list-group-item"><i></i><button class="close">&times;</button><span></span></li>');
		li[0].att = att;
		li.click(function() { setAttribute(att); });
		$('button', li).click(function() { removeAttribute(att); });
	}

	li.attr('title', att.description);
	li.children('span').html('&nbsp;' + att.name);
	
	// Create attribute list item
	var icon = 'unchecked';
	switch(att.type) {
		case 'string' :
			icon = 'align-justify';
			break;

		case 'group' :
			icon = 'chevron-right';
			break;
	}

	li.children('i').removeClass();
	li.children('i').addClass('glyphicon glyphicon-' + icon);

	return li;
}

function removeAttribute(att) {
	// find the attribute in citype
	for (i = 0; i < citype.attributes.length; i++) {
		if (att == citype.attributes[i]) {
			citype.attributes.splice(i, 1);
			break;
		}
	}

	li = getAttListItem(att);
	li.remove();
}

function setAttribute(att) {
	suspendUi = true;

	selectedAtt = att;

	// Select the matching list item
	liAtt = getAttListItem(att);

	// update the editor form
	inputAttName.val(att.name);
	inputAttDescription.val(att.description);
	selectAttType.val(att.type);

	// Update the active list item
	ulAtts.children('li').removeClass('active');
	liAtt.addClass('active');

	suspendUi = false;
}

function getAttListItem(att) {
	// select the correct list item
	var result = null;
	ulAtts.children('li').each(function(i, li) {
		if (li.att == att) {
			result = $(li);
			return;
		}
	});

	return result;
}

function updateCitype() {
	if (suspendUi) return;

	var att = selectedAtt;

	// Update the CI Type with form data
	citype.name = inputTypeName.val();
	citype.description = inputTypeDescription.val();

	// Update the attribute with form data
	if (att) {
		att.name = inputAttName.val();
		att.description = inputAttDescription.val();
		att.type = selectAttType.val();

		// Update the list item with the attribute data
		li = getAttListItem(att);
		buildAttributeListItem(att, li);
	}
}

function commitCitype() {
	submitData.val(JSON.stringify(citype));
    submitForm.submit();

	return false;
}

$(document).ready(function() {
	// Declare DOM elements
	ulAtts = $('#attlist');
	liNewAtt = $('#newatt');

	inputTypeName = $('#typeName');
	inputTypeDescription = $('#typeDesc');

	inputAttName = $('#attName');
	inputAttDescription = $('#attDesc');
	selectAttType = $('#attType');

	submitForm = $('#submitForm');
	submitData = $('#submitData');
	buttonSave = $('#save');

	// Wire up DOM events
	liNewAtt.click(addAttribute);

	inputTypeName.change(updateCitype);
	inputTypeDescription.change(updateCitype);
	inputAttName.change(updateCitype);
	inputAttDescription.change(updateCitype);
	selectAttType.change(updateCitype);
	
	buttonSave.click(commitCitype);
	
	
	// Display initial details
	for (i = 0; i < citype.attributes.length; i++) {
		att = citype.attributes[i];
		li = buildAttributeListItem(att, null);
		ulAtts.append(li);
	}

	if (citype.attributes.length > 0) {
		setAttribute(citype.attributes[0]);
	}
});