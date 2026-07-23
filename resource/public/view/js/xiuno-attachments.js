(function () {
  'use strict';

  function readAsDataURL(file) {
    return new Promise(function (resolve, reject) {
      var reader = new FileReader();
      reader.onload = function () { resolve(reader.result); };
      reader.onerror = function () { reject(new Error('读取附件失败')); };
      reader.readAsDataURL(file);
    });
  }

  function setStatus(widget, text, isError) {
    var status = widget.querySelector('.attachment-status');
    if (!status) return;
    status.textContent = text || '';
    status.classList.toggle('text-danger', !!isError);
    status.classList.toggle('text-muted', !isError);
  }

  function refreshList(widget) {
    var fieldset = widget.querySelector('.attachment-fieldset');
    if (!fieldset) return;
    fieldset.classList.toggle('d-none', !fieldset.querySelector('.attachment-item'));
  }

  function appendAttachment(widget, attachment) {
    var list = widget.querySelector('.attachlist');
    var item = document.createElement('li');
    item.className = 'attachment-item';
    item.dataset.aid = attachment.aid;

    var link = document.createElement('a');
    link.href = attachment.url;
    link.target = '_blank';
    link.rel = 'noopener';

    var icon = document.createElement('i');
    icon.className = 'icon filetype ' + (attachment.filetype || 'other');
    link.appendChild(icon);
    link.appendChild(document.createTextNode(' ' + attachment.orgfilename));

    var remove = document.createElement('a');
    remove.href = 'javascript:void(0)';
    remove.className = 'attachment-delete ml-3 text-danger';
    remove.innerHTML = '<i class="icon-remove"></i> 删除';

    item.appendChild(link);
    item.appendChild(remove);
    list.appendChild(item);
    refreshList(widget);
  }

  async function uploadOne(widget, file) {
    var data = await readAsDataURL(file);
    var body = new URLSearchParams();
    body.set('name', file.name || 'attach');
    body.set('data', data);
    body.set('width', '0');
    body.set('height', '0');
    // Xiuno 4.0.4's stock post uploader sends is_image=0 and lists the file
    // below the post. Image editor plugins can still opt into inline images.
    body.set('is_image', '0');

    var response = await fetch(widget.dataset.uploadUrl || '/attach/create', {
      method: 'POST',
      credentials: 'same-origin',
      headers: {'Content-Type': 'application/x-www-form-urlencoded;charset=UTF-8'},
      body: body.toString()
    });
    var payload = await response.json();
    if (String(payload.code) !== '0') {
      throw new Error(typeof payload.message === 'string' ? payload.message : '上传附件失败');
    }
    appendAttachment(widget, payload.message);
  }

  document.querySelectorAll('[data-attachment-widget]').forEach(function (widget) {
    var input = widget.querySelector('.attachment-input');
    var form = widget.closest('form');
    var submit = form ? form.querySelector('[type="submit"]') : null;
    refreshList(widget);

    input.addEventListener('change', async function () {
      var files = Array.prototype.slice.call(input.files || []);
      if (!files.length) return;
      if (submit) submit.disabled = true;
      try {
        for (var index = 0; index < files.length; index++) {
          setStatus(widget, '正在上传 ' + (index + 1) + '/' + files.length + '：' + files[index].name, false);
          await uploadOne(widget, files[index]);
        }
        setStatus(widget, '附件上传完成', false);
      } catch (error) {
        setStatus(widget, error.message || '上传附件失败', true);
      } finally {
        input.value = '';
        if (submit) submit.disabled = false;
      }
    });

    widget.addEventListener('click', async function (event) {
      var remove = event.target.closest('.attachment-delete');
      if (!remove || !widget.contains(remove)) return;
      event.preventDefault();
      var item = remove.closest('.attachment-item');
      if (!item || !window.confirm('确定删除这个附件吗？')) return;
      remove.classList.add('disabled');
      try {
        var response = await fetch('/attach/' + encodeURIComponent(item.dataset.aid) + '/delete', {
          method: 'POST', credentials: 'same-origin'
        });
        var payload = await response.json();
        if (String(payload.code) !== '0') {
          throw new Error(typeof payload.message === 'string' ? payload.message : '删除附件失败');
        }
        item.remove();
        refreshList(widget);
        setStatus(widget, '附件已删除', false);
      } catch (error) {
        remove.classList.remove('disabled');
        setStatus(widget, error.message || '删除附件失败', true);
      }
    });
  });
})();
