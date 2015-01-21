var attLookup = {};
var selectedAtt = null;

var attributeEditor = null;

var inputTypeName = null;
var inputTypeDescription = null;
var ulAtts = null;
var liNewAtt = null;
var inputAttName = null;
var inputAttDescription = null;
var selectAttType = null;
var inputAttRequired = null;

var editArray = null;
var inputAttMinCount = null;
var inputAttMaxCount = null;

var editNumber = null;
var inputNumberUnits = null;
var inputNumberMinValue = null;
var inputNumberMaxValue = null;

var editGroup = null;
var inputGroupSingular = null;

var editString = null;
var inputStringRequired = null;
var inputStringMinLength = null;
var inputStringMaxLength = null;

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
		if(a.children && a.children.length) {
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
	if (parent && ! parent.children)
		parent.children = [];

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
			iconClass = 'file-text-o';
			break;

		case 'number' :
			iconClass = 'sliders';
			break;

		case 'boolean' :
			iconClass = 'check-square-o';
			break;

		case 'group' :
			iconClass = 'chevron-right';
			break;
	}

	icon.removeClass().addClass('li-icon fa fa-' + iconClass);

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

	// Select parent attribute if the deleted attribute was selected
	if (att == selectedAtt) {
		if (att._parent) {
			setAttribute(att._parent);
		}

		else {
			if (citype.attributes && citype.attributes.length) {
				setAttribute(citype.attributes[0]);
			}

			else {
				setAttribute(undefined);
			}
		}
	}
}

function setAttribute(att) {
	suspendUi = true;

	// Set current attribute variables
	selectedAtt = att;
	attributeEditor.fadeOut(200, function() {
		if(! att)
			return;

		// update the editor form
		inputAttName.val(att.name);
		inputAttDescription.val(att.description);
		inputAttRequired.prop('checked', att.required ? true : false);
		selectAttType.val(att.type);

		// reset controls
		showControlGroup(att.type);

		// Update array editor
		inputAttArray.prop('checked', att.isArray ? true : false);
		inputAttMinCount.val(att.minCount);
		inputAttMaxCount.val(att.maxCount);
		
		switch(att.type) {
			case "string":
				inputStringMinLength.val(att.minLength);
				inputStringMaxLength.val(att.maxLength);
				break;

			case "group":
				if (att.isArray) {
					inputGroupSingular.val(att.singular);
					inputGroupSingular.show();
				}

				break;

			case "number":
				inputNumberUnits.val(att.units);
				inputNumberMinValue.val(att.minValue);
				inputNumberMaxValue.val(att.maxValue);
				break;
		}

		// Update the active list item
		$('li', ulAtts).removeClass('active');
		att._li.addClass('active');

		attributeEditor.fadeIn(200);
		suspendUi = false;
	});
}

function showControlGroup(attType) {

	// Reset array controls
	editArray.hide();
	inputAttArray.prop('checked', false);
	inputAttMinCount.val('');
	inputAttMaxCount.val('');

	// Reset string controls
	editString.hide();
	inputStringMinLength.val('');
	inputStringMaxLength.val('');

	// Reset number controls
	editNumber.hide();
	inputNumberUnits.val('');
	inputNumberMinValue.val('');
	inputNumberMaxValue.val('');

	// Reset group controls
	editGroup.hide();
	inputGroupSingular.hide();
	inputGroupSingular.val('');	

	switch(attType) {
		case "group":
			editArray.show();
			editGroup.show();
			break;

		case "string":
			editArray.show();
			editString.show();
			break;

		case "number":
			editArray.show();
			editNumber.show();
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
		att.required = inputAttRequired.is(':checked');

		// reset type specific attributes
		delete(att.isArray);
		delete(att.minCount);
		delete(att.maxCount);
		delete(att.singular);
		delete(att.minLength);
		delete(att.maxLength);
		delete(att.units);
		delete(att.minValue);
		delete(att.maxValue);

		switch(att.type) {
			case "group":
				att.isArray = inputAttArray.is(':checked');
				att.minCount = att.isArray ? parseInt(inputAttMinCount.val()) : undefined;
				att.maxCount = att.isArray ? parseInt(inputAttMaxCount.val()) : undefined;
				att.singular = att.isArray ? inputGroupSingular.val() : undefined;
				break;
			case "string":
				att.isArray = inputAttArray.is(':checked');
				att.minCount = att.isArray ? parseInt(inputAttMinCount.val()) : undefined;
				att.maxCount = att.isArray ? parseInt(inputAttMaxCount.val()) : undefined;
				att.minLength = parseInt(inputStringMinLength.val());
				att.maxLength = parseInt(inputStringMaxLength.val());
				break;
			case "number":
				att.isArray = inputAttArray.is(':checked');
				att.minCount = att.isArray ? parseInt(inputAttMinCount.val()) : undefined;
				att.maxCount = att.isArray ? parseInt(inputAttMaxCount.val()) : undefined;
				att.units = inputNumberUnits.val();
				att.minValue = parseFloat(inputNumberMinValue.val());
				att.maxValue = parseFloat(inputNumberMaxValue.val());
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

		if (a.children && a.children.length) {
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

	attributeEditor = $('#attributeEditor');

	inputTypeName = $('#typeName');
	inputTypeDescription = $('#typeDesc');

	inputAttName = $('#attName');
	inputAttDescription = $('#attDesc');
	inputAttRequired = $('#inputAttRequired');
	selectAttType = $('#attType');

	editArray = $('#editArray');
	inputAttArray = $('#inputAttArray');
	inputAttMinCount = $('#inputAttMinCount');
	inputAttMaxCount = $('#inputAttMaxCount');

	editGroup = $('#editGroup');
	inputGroupSingular = $('#inputGroupSingular');

	editString = $('#editString');
	inputStringRequired = $('#inputStringRequired');
	inputStringMinLength = $('#inputStringMinLength');
	inputStringMaxLength = $('#inputStringMaxLength');

	editNumber = $('#editNumber');
	inputNumberUnits = $('#inputNumberUnits');
	inputNumberMinValue = $('#inputNumberMinValue');
	inputNumberMaxValue = $('#inputNumberMaxValue');

	submitForm = $('#submitForm');
	submitData = $('#submitData');
	buttonSave = $('#save');

	// Wire up DOM events
	liNewAtt.click(function() { addAttribute(null); });
	inputTypeName.change(updateCitype);
	inputTypeDescription.change(updateCitype);
	inputAttName.change(updateCitype);
	inputAttDescription.change(updateCitype);
	inputAttRequired.change(updateCitype);
	inputAttMinCount.change(updateCitype);
	inputAttMaxCount.change(updateCitype);

	selectAttType.change(updateAttType);

	buttonSave.click(commitCitype);

	inputAttArray.change(function() {
		suspendUi = true;

		// Special consideration for showing group controls
		if(selectedAtt.type == "group") {
			if(inputAttArray.is(':checked')) {
				inputGroupSingular.show();
			} else {
				inputGroupSingular.hide();
				inputGroupSingular.val('');
			}
		}

		suspendUi = false;

		updateCitype();
		return false;
	});

	// String edit controls
	inputStringMinLength.change(updateCitype);
	inputStringMaxLength.change(updateCitype);

	// Number edit controls
	inputNumberUnits.change(updateCitype);
	inputNumberMinValue.change(updateCitype);
	inputNumberMaxValue.change(updateCitype);

	// Group edit controls
	inputGroupSingular.change(updateCitype);
	
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