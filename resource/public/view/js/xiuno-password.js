(function ($) {
	'use strict';

	$('form[data-xiuno-password-form]').on('submit', function () {
		var form = this;
		$(form).find('[data-xiuno-password-target]').each(function () {
			var target = form.elements[this.getAttribute('data-xiuno-password-target')];
			if (target) {
				target.value = this.value ? $.md5(this.value) : '';
			}
		});
	});
})(jQuery);
