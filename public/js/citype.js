var attLookup = {};
var selectedAtt = null;

var inputTypeName = null;
var inputTypeDescription = null;
var ulAtts = null;
var liNewAtt = null;
var inputAttName = null;
var inputAttDescription = null;
var selectAttType = null;

var editGroup = null;
var inputGroupArray = null;
var inputGroupSingular = null;

var submitForm = null;
var submitData = null;
var buttonSave = null;

var suspendUi = false;

// Build a hash table of attributes with contextual metadata
function buildMetadata(parent, path) {
	var atts = parent ? parent.children : citype.attributes;

	for (var i = 0; i < atts.length; i++) {
		a = atts[i];
		
		// Build a path
		var newPath = path + a.shortName;

		// Build metadata
		a._parent = parent;
		a._path = newPath;
		a._li = buildAttributeListItem(a);

		// Add attribute to hash table
		attLookup[newPath] = a;

		// Drill down into children
		if(a.children.length) {
			buildMetadata(a, newPath + '.');
		}
	}
}

function getShortName(name) {
	name = name.toLowerCase();
	name = name.replace(/ +/, '-');
	name = name.replace(/[^a-z0-9\-_]+/, '');
	name = name.replace(/-{2,}/,'-');

	return name;
}

function getPath(att) {
	for (var path in attLookup) {
		if (attLookup[path] == att) {
			return att._path;
		}
	}
	return null;
}

function addAttribute(parent) {
	var atts = parent ? parent.children : citype.attributes;

	// Find next available 'New Attribute' name
	var name = 'New Attribute';
	for (i = 1; i < 1024; i++) {
		var test = name;
		if (i > 1) test += ' ' + i;
		var matched = false;
		for (a = 0; a < atts.length; a++) {
			if (atts[a].name == test) {
				matched = true;
			}
		}

		if (!matched) {
			name = test;
			break;
		}
	}

	// Add a basic attribute
	var shortName = getShortName(name);
	var att = {
		"name" : name,
		"shortName" : shortName,
		"description": "",
		"type": "string",
		"children": [],
		"_parent": parent,
		"_path": parent ? parent._path + '.' + shortName : shortName
	};

	// Create a UI list item
	att._li = buildAttributeListItem(att, null);

	// Wire it in
	if(parent) {
		// Append to parent attribute
		parent.children.push(att);
		$('>ul.li-sublist', parent._li).append(att._li);
	} else {
		// Append to CI Type
		citype.attributes.push(att);
		ulAtts.append(att._li);
	}

	// Select the attribute in the UI
	setAttribute(att);

	// Send user focus to the attribute name field
	inputAttName.focus();
	inputAttName.select();
}

function buildAttributeListItem(att) {
	var container = $('> .li-container', att._li);
	var title = $('.li-title', container);
	var icon = $('.li-icon', container);
	var controls = $('.li-controls', container);
	var btnDelete = $('.li-delete', controls);
	var btnAdd = $('.li-add', controls);
	var ulChildren = $('>ul.li-sublist', att._li);

	if (! att._li) {
		// Create list item
		att._li = $('<li></li>');
		container = $('<span class="li-container link"></span>').appendTo(att._li);
		container.append($('<div class="clearfix"></div>'));
		icon = $('<i class="li-icon"></i>').appendTo(container);
		title = $('<span class="li-title"></span>').appendTo(container);
		controls = $('<span class="li-controls"></span>').appendTo(container);

		// Delete button
		btnDelete = $('<button class="close li-delete" data-toggle="tooltip" title="Delete">&times;</button>')
			.appendTo(controls);
		btnDelete.click(function() { removeAttribute(att); return false; });

		// Attach attribute to DOM
		att._li[0].att = att;

		// Wire up events
		container.click(function() { setAttribute(att); return false; });
	}

	// Update title
	title.html('&nbsp;' + att.name);
	
	// Update icon
	var iconClass = 'unchecked';
	switch(att.type) {
		case 'string' :
			iconClass = 'align-justify';
			break;

		case 'group' :
			iconClass = 'chevron-right';
			break;
	}

	icon.removeClass().addClass('li-icon glyphicon glyphicon-' + iconClass);

	// Add/remove 'add child' button
	if (att.type == 'group') {
		if (! btnAdd.length) {
			btnAdd = $('<button class="btn li-add" data-toggle="tooltip" data-original-title="Add child">&plus;</button>')
				.appendTo($('span.li-controls', att._li));
			btnAdd.click(function() { addAttribute(att); return false; });
		}

		if (! ulChildren.length) {
			ulChildren = $('<ul class="li-sublist list-unstyled"></ul>')
				.appendTo(att._li);
		}
	} else {
		// Ensure group controls are cleaned up
		btnAdd.remove();
		ulChildren.remove();
	}

	return att._li;
}

function removeAttribute(att) {
	// remove attribute from parent
	var atts = att._parent ? att._parent.children : citype.attributes;
	for (i = 0; i < atts.length; i++) {
		if (att == atts[i]) {
			atts.splice(i, 1);
			break;
		}
	}

	// remove list item from ui
	att._li.remove();
}

function setAttribute(att) {
	suspendUi = true;

	// Set current attribute variables
	selectedAtt = att;

	// update the editor form
	inputAttName.val(att.name);
	inputAttDescription.val(att.description);
	selectAttType.val(att.type);

	showControlGroup(att.type);
	
	switch(att.type) {
		case "string":
			break;

		case "group":
			if (att.isArray) {
				inputGroupArray.prop('checked', true);
				inputGroupSingular.val(att.singular);
				inputGroupSingular.show();
			}

			break;
	}

	// Update the active list item
	$('li', ulAtts).removeClass('active');
	att._li.addClass('active');

	suspendUi = false;
}

function showControlGroup(attType) {

	// Reset edit controls
	editString.hide();

	editGroup.hide();
	inputGroupSingular.hide();
	inputGroupArray.prop('checked', false);
	inputGroupSingular.val('');	

	switch(attType) {
		case "group":
			editGroup.show();
			break;
	}
}

function updateCitype() {
	if (suspendUi) return;

	var att = selectedAtt;

	// Update the CI Type with form data
	citype.name = inputTypeName.val();
	citype.description = inputTypeDescription.val();

	// Update the selected attribute with form data
	if (att) {
		att.name = inputAttName.val();
		att.description = inputAttDescription.val();
		att.type = selectAttType.val();

		switch(att.type) {
			case "group":
				att.isArray = inputGroupArray.is(':checked');
				att.singular = att.isArray ? inputGroupSingular.val() : null;
				break;
		}

		// Update the list item with the attribute data
		buildAttributeListItem(att);
	}
}

function updateAttType() {
	var type = selectAttType.val();

	if (type != 'group') {
		// remove any children if it's not a group
		selectedAtt.children = [];
		$('ul', selectedAtt._li).remove();
	}

	showControlGroup(type);

	updateCitype();
}

function commitCitype() {
	// Filter out ._* metadata attributes
	var filter = function(key, val) {
		if (key.match(/^_/))
			return undefined;
		return val;
	};

	// Convert citype to JSON string
	submitData.val(JSON.stringify(citype, filter));

	// To the cloud
    submitForm.submit();

	return false;
}

function buildAttTree(attributes, list, path) {
	for (var i = 0; i < attributes.length; i++) {
		a = attributes[i];
		list.append(a._li);

		if (a.children.length) {
			var ul = $('ul.li-sublist', a._li);

			buildAttTree(a.children, ul, path + a.shortName + '.');
		}
	}
}

$(document).ready(function() {
	// Init CI Type data
	buildMetadata(null, '');

	// Declare DOM elements
	ulAtts = $('#attlist');
	liNewAtt = $('#newatt');

	inputTypeName = $('#typeName');
	inputTypeDescription = $('#typeDesc');

	inputAttName = $('#attName');
	inputAttDescription = $('#attDesc');
	selectAttType = $('#attType');

	editGroup = $('#editGroup');
	inputGroupArray = $('#inputGroupArray');
	inputGroupSingular = $('#inputGroupSingular');

	editString = $('#editString');

	submitForm = $('#submitForm');
	submitData = $('#submitData');
	buttonSave = $('#save');

	// Wire up DOM events
	liNewAtt.click(function() { addAttribute(null); });
	inputTypeName.change(updateCitype);
	inputTypeDescription.change(updateCitype);
	inputAttName.change(updateCitype);
	inputAttDescription.change(updateCitype);
	selectAttType.change(updateAttType);
	
	inputGroupArray.change(function() {
		suspendUi = true;
		if(inputGroupArray.is(':checked')) {
			inputGroupSingular.show();
		} else {
			inputGroupSingular.hide();
			inputGroupSingular.val('');
		}
		suspendUi = false;

		updateCitype();
		return false;
	});
	inputGroupSingular.change(updateCitype);

	buttonSave.click(commitCitype);
	
	// Display initial details
	buildAttTree(citype.attributes, ulAtts, '');

	// Select first or create an attribute
	if (citype.attributes.length === 0) {
		addAttribute(null);
	} else {
		setAttribute(citype.attributes[0]);
	}

	// Init tooltips
	$('[data-toggle="tooltip"]').tooltip();
});