<script setup>
import {ref} from 'vue'
import axios from 'axios'

const emit = defineEmits(['close', 'reload'])

const props = defineProps(['story_id'])

const file = ref(null)
const title = ref("")
const description = ref("")
const fileMsg = ref("Upload file")

const handleFileUpload = (event) => {
    if (event.target.files[0].size > 500*1024) {
        fileMsg.value = "File is too large"
        return
    }
    file.value = event.target.files[0]
}

function getCookie(name) {
 const value = `; ${document.cookie}`;
 const parts = value.split(`; ${name}=`);
 if (parts.length === 2) return parts.pop().split(';').shift();
}

const refreshTokens = async()=>{
    try {
        let {data} = await axios.get('http://localhost:3001/api/users/refresh', {
            headers:{
                'Refresh': getCookie('Refresh'),
            }
        })

        document.cookie = `Authorization=${data.authorization};`
        document.cookie = `Refresh=${data.refresh};`
    } catch (error) {
        if (error.response.status == 401){
            await refreshTokens()
            loadContent()
        }
        else console.log(error)
    }
}

const createBanner = async ()=>{
    const formData = new FormData()
    formData.append('file', file.value)
    formData.append('name', title.value)
    formData.append('description', description.value) 
    console.log(getCookie("Authorization"))
    try{
        let url = "http://localhost:3001/api/banners"
        if (props.story_id) url += `?story_id=${props.story_id}`
        await axios.post(url, formData, {
            headers: {
                'Content-Type':'multipart/form-data',
                'Authorization': getCookie("Authorization")
            }
        })
        emit('close')
        emit('reload')
    } catch (error) {
        console.log(error)
    }
}




</script>

<template>
    <div @click.self="$emit('close')" class="modal-wrapper w-screen h-screen absolute z-50 top-0 left-0 backdrop-blur-sm flex justify-center items-center">
        <div class="modal flex flex-col bg-gray-900 text-white p-5 rounded-lg items-center gap-2">
            <h2 class="title">New banner</h2>
            <input v-model="title" type="text" class="text-input" placeholder="Title">
            <textarea v-model="description" type="text" class="text-input" placeholder="Description" rows="5"></textarea>
    
            <input type="file" id="fileInput" class="hidden" @change="handleFileUpload" />
            <label for="fileInput" class="button text-center w-72">
                <p class="truncate">{{ file != null ? file.name : fileMsg }}</p>
            </label>
    
            <button @click="createBanner" class="button">Create</button>
        </div>
    </div>


</template>